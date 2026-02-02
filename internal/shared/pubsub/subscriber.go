package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hellotect2022go/chatapp/internal/shared/model"
	"github.com/hellotect2022go/chatapp/internal/websocket/hub"
	"github.com/redis/go-redis/v9"
)

type Subscriber struct {
	rdb     *redis.Client
	roomHub *hub.RoomHub // ⭐ WebSocket 레이어와 연결
	ctx     context.Context
}

func NewSubscriber(rdb *redis.Client, roomHub *hub.RoomHub) *Subscriber {
	return &Subscriber{
		rdb:     rdb,
		roomHub: roomHub,
		ctx:     context.Background(),
	}
}

// Redis 구독 시작 (기존 SubscribeAll과 동일)
func (s *Subscriber) SubscribeAll() {
	pubsub := s.rdb.PSubscribe(s.ctx, "chat:room:*", "room:event")
	defer pubsub.Close()

	log.Println("Redis Subscribe started: chat:room:*, room:event")

	ch := pubsub.Channel()
	for msg := range ch {
		if strings.HasPrefix(msg.Channel, "chat:room:") {
			// 채팅 메시지 처리
			var roomID uint
			fmt.Sscanf(msg.Channel, "chat:room:%d", &roomID)

			s.roomHub.Broadcast <- &hub.BroadcastMessage{
				RoomID:  roomID,
				Message: []byte(msg.Payload),
			}

		} else if msg.Channel == "room:event" {
			// 방 이벤트 처리
			var event map[string]interface{}
			if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
				log.Printf("Failed to parse room event: %v", err)
				continue
			}

			log.Println("room:event", event)
			s.handleRoomEvent(event)
		}
	}
}

// 방 이벤트 핸들러 (기존 handleRoomEvent와 동일)
func (s *Subscriber) handleRoomEvent(event map[string]interface{}) {
	eventType, ok := event["type"].(string)
	if !ok {
		log.Printf("Invalid event type")
		return
	}

	switch eventType {
	case "room_created":
		roomID := uint(event["room_id"].(float64))
		members := event["members"].([]interface{})

		for _, memberInterface := range members {
			userID := uint(memberInterface.(float64))
			s.roomHub.JoinUserToRoom(userID, roomID)
		}

	case "member_added":
		roomID := uint(event["room_id"].(float64))
		userID := uint(event["user_id"].(float64))
		s.roomHub.JoinUserToRoom(userID, roomID)

	case "member_removed":
		roomID := uint(event["room_id"].(float64))
		userID := uint(event["user_id"].(float64))
		s.roomHub.RemoveUserFromRoom(userID, roomID)

	case "user_joined":
		roomID := uint(event["room_id"].(float64))
		userID := uint(event["user_id"].(float64))
		nickname := event["nickname"].(string)

		joinMsg := &model.WSMessage{
			Type:      model.MessageTypeJoin,
			RoomID:    roomID,
			UserID:    userID,
			Nickname:  nickname,
			Content:   fmt.Sprintf("%s님이 입장하셨습니다.", nickname),
			Timestamp: time.Now(),
		}
		msgBytes, _ := json.Marshal(joinMsg)
		s.roomHub.Broadcast <- &hub.BroadcastMessage{
			RoomID:  roomID,
			Message: msgBytes,
		}

	default:
		log.Printf("Unknown event type: %s", eventType)
	}
}
