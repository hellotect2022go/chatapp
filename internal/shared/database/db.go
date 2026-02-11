package database

import (
	"fmt"
	"os"
	"time"

	"github.com/hellotect2022go/chatapp/internal/shared/logger"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
)

func ConnectDB() *gorm.DB {
	// 1. .env 파일 로드
	err := godotenv.Load()
	if err != nil {
		logger.Warn(".env file not found, using system environment variables")
	}

	// 2. 환경 변수에서 값 가져오기
	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASSWORD")
	name := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")

	// 3. DSN 조립
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		host, user, pass, name, port)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger:                                   glogger.Default.LogMode(glogger.Info),
	})
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	logger.Info("Database connected successfully!")

	// 연결 풀 설정
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(100)                   // 최대 연결 수
	sqlDB.SetConnMaxIdleTime(10)                 // 최대 유휴 연결 수
	sqlDB.SetConnMaxLifetime(3600 * time.Second) // 연결 최대 수명(초)

	return db
}
