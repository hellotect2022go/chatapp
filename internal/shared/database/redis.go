package database

import (
	"context"
	"os"
	"strconv"

	"github.com/hellotect2022go/chatapp/internal/shared/logger"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func NewRedisClient() *redis.Client {
	// 1. .env 파일 로드 (이미 main이나 다른 곳에서 로드했다면 생략 가능)
	if err := godotenv.Load(); err != nil {
		logger.Warn(".env file not found, using system environment variables")
	}

	// 2. 환경 변수 읽기
	addr := os.Getenv("REDIS_ADDR")
	password := os.Getenv("REDIS_PASSWORD")
	dbStr := os.Getenv("REDIS_DB")

	// 3. DB 번호는 int 타입이므로 변환 필요
	dbNum, _ := strconv.Atoi(dbStr)

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       dbNum,
	})

	// 5. 연결 확인 (Ping)
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Fatal("Failed to connect to Redis", zap.Error(err))
	}

	logger.Info("Redis connected successfully", zap.String("addr", addr))
	return rdb
}
