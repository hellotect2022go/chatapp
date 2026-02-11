package hub

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
)

type RoomHub struct {
	rooms map[string]map[*Client]bool // ⭐ room_id(string) -> clients
	users map[string]*Client           // ⭐ user_uid(string) -> client

	Join      chan *JoinRequest
	Leave     chan *LeaveRequest
	Broadcast chan *BroadcastMessage

	mu       sync.RWMutex
	shutdown chan struct{} // ⭐ 종료 시그널
}

type JoinRequest struct {
	RoomID string // ⭐ uint → string
	Client *Client
}

type LeaveRequest struct {
	RoomID string // ⭐ uint → string
	Client *Client
}

type BroadcastMessage struct {
	RoomID  string // ⭐ uint → string
	Message []byte
}

func NewRoomHub() *RoomHub {
	return &RoomHub{
		rooms:     make(map[string]map[*Client]bool),
		users:     make(map[string]*Client),
		Join:      make(chan *JoinRequest),
		Leave:     make(chan *LeaveRequest),
		Broadcast: make(chan *BroadcastMessage),
		shutdown:  make(chan struct{}),
	}
}

func (h *RoomHub) Run() {
	log.Println("RoomHub started")
	for {
		select {
		case <-h.shutdown:
			log.Println("RoomHub shutting down...")
			h.closeAllConnections()
			return

		case req := <-h.Join:
			h.handleJoin(req)

		case req := <-h.Leave:
			h.handleLeave(req)

		case msg := <-h.Broadcast:
			h.handleBroadcast(msg)
		}
	}
}

func (h *RoomHub) handleJoin(req *JoinRequest) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.rooms[req.RoomID]; !ok {
		h.rooms[req.RoomID] = make(map[*Client]bool)
	}

	h.rooms[req.RoomID][req.Client] = true
	h.users[req.Client.UserUID] = req.Client
	req.Client.AddRoom(req.RoomID)

	log.Printf("User %s joined room %s (WebSocket connection)", req.Client.UserUID, req.RoomID)
	// ⭐ 자동 입장 메시지 제거: API에서 처리
}

func (h *RoomHub) handleLeave(req *LeaveRequest) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, ok := h.rooms[req.RoomID]; ok {
		if _, ok := clients[req.Client]; ok {
			// ⭐ 자동 퇴장 메시지 제거: API에서 처리
			delete(clients, req.Client)
			req.Client.RemoveRoom(req.RoomID)

			log.Printf("User %s left room %s (WebSocket disconnection)", req.Client.UserUID, req.RoomID)

			if len(clients) == 0 {
				delete(h.rooms, req.RoomID)
			}
		}
	}
}

func (h *RoomHub) broadcastJoinLeaveEvent(roomID string, userUID string, nickname string, eventType string) {
	var content string
	if eventType == "join" {
		content = fmt.Sprintf("%s님이 입장했습니다.", nickname)
	} else if eventType == "leave" {
		content = fmt.Sprintf("%s님이 퇴장했습니다.", nickname)
	}

	event := map[string]interface{}{
		"type":      eventType,
		"room_id":   roomID,
		"user_uid":  userUID,
		"nickname":  nickname,
		"content":   content,
		"timestamp": time.Now(),
	}

	eventBytes, _ := json.Marshal(event)

	if clients, ok := h.rooms[roomID]; ok {
		for client := range clients {
			select {
			case client.Send <- eventBytes:
			default:
			}
		}
	}
}

func (h *RoomHub) handleBroadcast(msg *BroadcastMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if clients, ok := h.rooms[msg.RoomID]; ok {
		for client := range clients {
			select {
			case client.Send <- msg.Message:
			default:
				close(client.Send)
				delete(clients, client)
				delete(h.users, client.UserUID)
			}
		}
	}
}

// ⭐ 추가: UserUID로 특정 방에 동적으로 참여시키기
func (h *RoomHub) JoinUserToRoom(userUID string, roomID string) bool {
	h.mu.RLock()
	client, exists := h.users[userUID]
	h.mu.RUnlock()

	if !exists {
		log.Printf("User %s not connected, cannot join room %s", userUID, roomID)
		return false
	}

	// 이미 참여한 방인지 확인
	if client.IsInRoom(roomID) {
		log.Printf("User %s already in room %s", userUID, roomID)
		return false
	}

	// Join 요청
	h.Join <- &JoinRequest{
		RoomID: roomID,
		Client: client,
	}

	log.Printf("User %s dynamically joined room %s", userUID, roomID)
	return true
}

// ⭐ 추가: UserUID로 특정 방에서 나가기
func (h *RoomHub) RemoveUserFromRoom(userUID string, roomID string) bool {
	h.mu.RLock()
	client, exists := h.users[userUID]
	h.mu.RUnlock()

	if !exists {
		return false
	}

	if !client.IsInRoom(roomID) {
		return false
	}

	h.Leave <- &LeaveRequest{
		RoomID: roomID,
		Client: client,
	}

	log.Printf("User %s dynamically left room %s", userUID, roomID)
	return true
}

// ⭐ 추가: Client 완전 제거 (연결 종료 시)
func (h *RoomHub) RemoveClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 모든 방에서 제거
	for roomID := range client.Rooms {
		if clients := h.rooms[roomID]; clients != nil {
			delete(clients, client)
			if len(clients) == 0 {
				delete(h.rooms, roomID)
			}
		}
	}

	// UserUID 매핑 제거
	delete(h.users, client.UserUID)

	log.Printf("User %s completely removed from hub", client.UserUID)
}

// ⭐ Shutdown - RoomHub 종료
func (h *RoomHub) Shutdown() {
	close(h.shutdown)
}

// ⭐ closeAllConnections - 모든 클라이언트 연결 종료
func (h *RoomHub) closeAllConnections() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, clients := range h.rooms {
		for client := range clients {
			close(client.Send)
			client.Conn.Close()
		}
	}

	h.rooms = make(map[string]map[*Client]bool)
	h.users = make(map[string]*Client)

	log.Println("All client connections closed")
}

// ⭐ GetClientCount - 전체 클라이언트 수
func (h *RoomHub) GetClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.users)
}

// ⭐ GetRoomCount - 전체 방 수
func (h *RoomHub) GetRoomCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.rooms)
}
