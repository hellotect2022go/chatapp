package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/hellotect2022go/chatapp/internal/api/middleware"
	"github.com/hellotect2022go/chatapp/internal/shared/logger"
	"github.com/hellotect2022go/chatapp/internal/websocket/infra"
	"go.uber.org/zap"
)

func main() {
	// 로거 초기화
	mode := logger.GetMode()
	logger.InitLogger(mode)
	defer logger.Sync()

	logger.Info("WebSocket Server Starting",
		zap.String("mode", mode),
		zap.String("port", "9999"),
		zap.String("service", "websocket"))

	// Container 생성
	app := infra.NewWSContainer()
	defer app.Close()

	// Router 설정
	r := setupRouter(app)

	// Graceful Shutdown 설정
	go func() {
		logger.Info("WebSocket Server listening on :9999")
		if err := r.Run(":9999"); err != nil {
			logger.Fatal("WebSocket Server failed to start", zap.Error(err))
		}
	}()

	// 종료 시그널 대기
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down WebSocket Server...")
}

func setupRouter(app *infra.WSContainer) *gin.Engine {
	r := gin.Default()

	// 미들웨어 (최소한만)
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS())

	// Health Check
	r.GET("/health", func(c *gin.Context) {
		stats := app.GetStats() // Container에서 통계 제공
		c.JSON(200, gin.H{
			"status":        "ok",
			"service":       "websocket",
			"total_clients": stats.TotalClients,
			"total_rooms":   stats.TotalRooms,
		})
	})

	// Readiness Check (K8s용)
	// r.GET("/readiness", func(c *gin.Context) {
	// 	if app.IsReady() {
	// 		c.JSON(200, gin.H{"status": "ready"})
	// 	} else {
	// 		c.JSON(503, gin.H{"status": "not ready"})
	// 	}
	// })

	// ===== WebSocket 엔드포인트 =====
	r.GET("/ws", app.ConnectionHandler.HandleConnection)

	return r
}
