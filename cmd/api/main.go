package main

import (
	"github.com/hellotect2022go/chatapp/internal/api/infra"
	"github.com/hellotect2022go/chatapp/internal/shared/logger"
	"go.uber.org/zap"
)

func initLogger() string {
	mode := logger.GetMode() // 로거 모드 가져오기
	logger.InitLogger(mode)
	logger.Info("Application Starting", zap.String("mode", mode), zap.String("type", "api"))
	return mode
}

func main() {

	// 1. 환경 설정 및 로거 초기화
	mode := logger.GetMode() // 로거 모드 가져오기
	logger.InitLogger(mode)
	defer logger.Sync()
	logger.Info("Application Starting", zap.String("mode", mode), zap.String("type", "api"))

	// 2. 컨테이너 생성(의존성 조립)
	app := infra.NewContainer()
	// 3. Router 설정
	router := app.SetupRouter()

	logger.Info("Server starting on :8888", zap.String("mode", mode))
	if err := router.Run(":8888"); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
