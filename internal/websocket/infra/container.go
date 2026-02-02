package infra

import (
	"sync"

	"github.com/hellotect2022go/chatapp/internal/shared/database"
	"github.com/hellotect2022go/chatapp/internal/shared/logger"
	"github.com/hellotect2022go/chatapp/internal/shared/pubsub"
	"github.com/hellotect2022go/chatapp/internal/websocket/handler"
	"github.com/hellotect2022go/chatapp/internal/websocket/hub"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type WSContainer struct {
	Redis             *redis.Client
	Publisher         *pubsub.Publisher
	Subscriber        *pubsub.Subscriber
	RoomHub           *hub.RoomHub
	ConnectionHandler *handler.ConnectionHandler

	// Graceful shutdown용
	done chan struct{}
	wg   sync.WaitGroup
}

func NewWSContainer() *WSContainer {
	// 1. Redis 연결
	rdb := database.NewRedisClient()
	logger.Info("WebSocket - Redis connected")

	// 2. Pub/Sub 초기화
	publisher := pubsub.NewPublisher(rdb)
	logger.Info("WebSocket - Publisher initialized")

	// 3. RoomHub 초기화
	roomHub := hub.NewRoomHub()
	logger.Info("WebSocket - RoomHub initialized")

	// 4. Subscriber 초기화
	subscriber := pubsub.NewSubscriber(rdb, roomHub)
	logger.Info("WebSocket - Subscriber initialized")

	// 5. Connection Handler 초기화
	connectionHandler := handler.NewConnectionHandler(roomHub, publisher, rdb)
	logger.Info("WebSocket - ConnectionHandler initialized")

	container := &WSContainer{
		Redis:             rdb,
		Publisher:         publisher,
		Subscriber:        subscriber,
		RoomHub:           roomHub,
		ConnectionHandler: connectionHandler,
		done:              make(chan struct{}),
	}

	// 6. 백그라운드 서비스 시작
	container.startBackgroundServices()

	return container
}

// startBackgroundServices - 백그라운드 고루틴 시작
func (c *WSContainer) startBackgroundServices() {
	// RoomHub 실행
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		c.RoomHub.Run()
	}()
	logger.Info("WebSocket - RoomHub started")

	// Subscriber 실행
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		c.Subscriber.SubscribeAll()
	}()
	logger.Info("WebSocket - Subscriber started")
}

// Close - 리소스 정리
func (c *WSContainer) Close() {
	logger.Info("WebSocket - Closing container...")

	// 백그라운드 서비스 종료
	close(c.done)

	// RoomHub 종료 (모든 클라이언트 연결 종료)
	c.RoomHub.Shutdown()

	// 고루틴 종료 대기
	c.wg.Wait()

	// Redis 연결 종료
	if err := c.Redis.Close(); err != nil {
		logger.Error("Failed to close Redis", zap.Error(err))
	}

	logger.Info("WebSocket - Container closed")
}

// GetStats - 통계 정보 반환
func (c *WSContainer) GetStats() struct {
	TotalClients int
	TotalRooms   int
} {
	return struct {
		TotalClients int
		TotalRooms   int
	}{
		TotalClients: c.RoomHub.GetClientCount(),
		TotalRooms:   c.RoomHub.GetRoomCount(),
	}
}
