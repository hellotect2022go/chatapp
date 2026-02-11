package pubsub

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hellotect2022go/chatapp/internal/shared/model"
	"github.com/redis/go-redis/v9"
)

type Publisher struct {
	rdb *redis.Client
	ctx context.Context
}

func NewPublisher(rdb *redis.Client) *Publisher {
	return &Publisher{
		rdb: rdb,
		ctx: context.Background(),
	}
}

// 채널명 생성
func (p *Publisher) channelName(roomID string) string {
	return fmt.Sprintf("chat:room:%s", roomID) // ⭐ string 기반
}

// 1. 일반 메시지 발행
func (p *Publisher) Publish(roomID string, message []byte) error {
	channel := p.channelName(roomID)
	return p.rdb.Publish(p.ctx, channel, message).Err()
}

// 2. WSMessage 구조체 발행
func (p *Publisher) PublishMessage(roomID string, msg *model.WSMessage) error {
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return p.Publish(roomID, msgBytes)
}

// 3. 방 이벤트 발행 (room:event 채널)
func (p *Publisher) PublishRoomEvent(event map[string]interface{}) error {
	eventBytes, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return p.rdb.Publish(p.ctx, "room:event", eventBytes).Err()
}
