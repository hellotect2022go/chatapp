package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	ws "github.com/gorilla/websocket"
	"github.com/hellotect2022go/chatapp/internal/shared/model"
	"github.com/hellotect2022go/chatapp/internal/shared/pubsub"
	"github.com/hellotect2022go/chatapp/internal/shared/util"
	"github.com/hellotect2022go/chatapp/internal/websocket/hub"
	"github.com/redis/go-redis/v9"
)

const (
	writeWait      = 10 * time.Second    // 쓰기 대기시간
	pongWait       = 60 * time.Second    // Pong 대기시간
	pingPeriod     = (pongWait * 9) / 10 // Ping 대기시간
	maxMessageSize = 512                 // 최대 메세지 사이즈
)

type ConnectionHandler struct {
	roomHub   *hub.RoomHub
	publisher *pubsub.Publisher
	rdb       *redis.Client
	upgrader  ws.Upgrader
}

func NewConnectionHandler(
	roomHub *hub.RoomHub,
	publisher *pubsub.Publisher,
	rdb *redis.Client,
) *ConnectionHandler {

	return &ConnectionHandler{
		roomHub:   roomHub,
		publisher: publisher,
		rdb:       rdb,
		upgrader: ws.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // TODO: 프로덕션에서는 제한 필요
			},
		},
	}
}

// HandleConnection - WebSocket 연결 핸들러
func (h *ConnectionHandler) HandleConnection(c *gin.Context) {
	// 1. 토큰 검증
	token := c.Query("token")
	if token == "" {
		c.JSON(401, gin.H{"error": "Token required in query parameter"})
		return
	}

	claims, err := util.ValidateToken(token)
	if err != nil {
		c.JSON(401, gin.H{"error": "Invalid or expired token"})
		return
	}

	// 2. WebSocket Upgrade
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	// 3. 클라이언트 생성 (UID 기반)
	client := hub.NewClient(h.roomHub, conn, claims.UserUID, claims.Nickname)

	// 4. Redis에서 사용자의 채팅방 목록 조회 (재연결 대응)
	h.autoJoinRooms(client)

	// 5. Read/Write 고루틴 시작
	go h.handleClientWrite(client)
	go h.handleClientRead(client)

	log.Printf("User %s connected (nickname: %s)", client.UserUID, client.Nickname)
}

// autoJoinRooms - 사용자의 채팅방 자동 입장
func (h *ConnectionHandler) autoJoinRooms(client *hub.Client) {
	ctx := context.Background()
	userRoomsKey := fmt.Sprintf("user:%s:rooms", client.UserUID) // ⭐ UID 기반 키

	roomIDs, err := h.rdb.SMembers(ctx, userRoomsKey).Result()
	if err != nil {
		log.Printf("Failed to get user rooms from Redis: %v", err)
		return
	}

	for _, roomID := range roomIDs {
		h.roomHub.Join <- &hub.JoinRequest{
			RoomID: roomID, // ⭐ 이미 string
			Client: client,
		}
		log.Printf("User %s auto-joined room %s", client.UserUID, roomID)
	}
}

// handleClientRead - 클라이언트로부터 메시지 읽기
func (h *ConnectionHandler) handleClientRead(client *hub.Client) {
	defer func() {
		h.roomHub.RemoveClient(client)
		client.Conn.Close()
		log.Printf("User %s disconnected", client.UserUID)
	}()

	// 타임아웃 설정
	client.Conn.SetReadDeadline(time.Now().Add(pongWait))
	client.Conn.SetReadLimit(maxMessageSize)
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			if ws.IsUnexpectedCloseError(err, ws.CloseGoingAway, ws.CloseAbnormalClosure) {
				log.Printf("Unexpected close error: %v", err)
			}
			break
		}

		// 메시지 처리
		h.handleMessage(client, message)
	}
}

// handleMessage - 수신한 메시지 처리
func (h *ConnectionHandler) handleMessage(client *hub.Client, message []byte) {
	var wsMsg model.WSMessage
	if err := json.Unmarshal(message, &wsMsg); err != nil {
		log.Printf("Invalid message format: %v", err)
		return
	}

	// ⭐ WebSocket은 실시간 알림만 처리
	switch wsMsg.Type {
	case model.MessageTypeTyping:
		// ✅ 타이핑 알림만 허용 (DB 저장 불필요)
		h.handleTypingNotification(client, &wsMsg)

	case model.MessageTypeChat, model.MessageTypeJoin, model.MessageTypeLeave:
		// ❌ API로 처리해야 함
		errMsg := gin.H{
			"type":    "error",
			"message": "Use API endpoints for this action",
			"details": "POST /api/v1/rooms/:id/messages for chat, POST /api/v1/rooms/:id/join for join",
		}
		msgBytes, _ := json.Marshal(errMsg)
		client.Send <- msgBytes

	default:
		log.Printf("Unknown message type: %s", wsMsg.Type)
	}
}

// handleTypingNotification - 타이핑 알림 처리
func (h *ConnectionHandler) handleTypingNotification(client *hub.Client, wsMsg *model.WSMessage) {
	wsMsg.UserUID = client.UserUID // ⭐ UserID → UserUID
	wsMsg.Nickname = client.Nickname
	wsMsg.Timestamp = time.Now()

	msgBytes, err := json.Marshal(wsMsg)
	if err != nil {
		log.Printf("Failed to marshal typing notification: %v", err)
		return
	}

	// Redis Pub/Sub으로 발행
	if err := h.publisher.Publish(wsMsg.RoomID, msgBytes); err != nil {
		log.Printf("Redis publish error for typing: %v", err)
	}
}

// handleClientWrite - 클라이언트로 메시지 전송
func (h *ConnectionHandler) handleClientWrite(client *hub.Client) {
	ticker := time.NewTicker(pingPeriod) // 일정 시간 이후 ticker.Channel 에 신호 보냄
	defer func() {
		ticker.Stop()
		client.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			client.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub가 채널을 닫음
				client.Conn.WriteMessage(ws.CloseMessage, []byte{})
				return
			}

			w, err := client.Conn.NextWriter(ws.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// 큐에 있는 메시지 모두 전송
			n := len(client.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-client.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C: // ticker.Channel 에서 값이 들어오면 ping 날림
			client.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := client.Conn.WriteMessage(ws.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
