package infra

import (
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/hellotect2022go/chatapp/internal/api/handler"
	"github.com/hellotect2022go/chatapp/internal/api/middleware"
	"github.com/hellotect2022go/chatapp/internal/api/repository"
	"github.com/hellotect2022go/chatapp/internal/api/service"
	"github.com/hellotect2022go/chatapp/internal/shared/database"
	"github.com/hellotect2022go/chatapp/internal/shared/logger"
	"github.com/hellotect2022go/chatapp/internal/shared/model"
	"github.com/hellotect2022go/chatapp/internal/shared/pubsub"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Container struct {
	DB              *gorm.DB
	Redis           *redis.Client
	AuthHandler     *handler.AuthHandler
	RoomHandler     *handler.RoomHandler
	MessageHandler  *handler.MessageHandler
	FileHandler     *handler.FileHandler
	UserHandler     *handler.UserHandler
	RecoveryHandler *handler.RedisRecoveryHandler
}

func NewContainer() *Container {
	// 1. DB 연결 및 마이그레이션
	db := database.ConnectDB()
	migrateDB(db)

	// 2. Redis 연결
	rdb := database.NewRedisClient()
	publisher := pubsub.NewPublisher(rdb)

	// 3. (나중에 추가될 부분) Repo -> Service -> Handler 조립
	userRepo := repository.NewUserRepository(db, rdb)
	sessionRepo := repository.NewSessionRepository(rdb)
	roomRepo := repository.NewRoomRepository(db)
	messageRepo := repository.NewMessageRepository(db)
	fileRepo := repository.NewFileRepository(db)

	authService := service.NewAuthService(userRepo, sessionRepo)
	roomService := service.NewRoomService(roomRepo, messageRepo, userRepo, rdb, publisher)
	messageService := service.NewMessageService(messageRepo, roomRepo, userRepo, publisher)
	recoveryService := service.NewRedisRecoveryService(rdb, roomRepo)
	userService := service.NewUserService(userRepo)
	fileService := service.NewFileService(fileRepo, "http://localhost:8888")

	healthy, err := recoveryService.CheckRedisHealth()
	if err != nil {
		logger.Warn("Redis health check error", zap.Error(err))
	} else if !healthy {
		logger.Warn("Redis data not found - starting recovery...")
		if err := recoveryService.RecoverUserRoomFromDB(); err != nil {
			logger.Error("Redis recovery failed", zap.Error(err))
		} else {
			logger.Info("Redis recovery completed successfully")
		}
	}

	// 7. Handler
	authHandler := handler.NewAuthHandler(authService)
	roomHandler := handler.NewRoomHandler(roomService, messageService) // ⭐ MessageService 추가
	messageHandler := handler.NewMessageHandler(messageService, fileService)
	fileHandler := handler.NewFileHandler(fileService) // ⭐ 추가: 파일 핸들러
	userHandler := handler.NewUserHandler(userService) // ⭐ 추가: 사용자 핸들러
	recoveryHandler := handler.NewRedisRecoveryHandler(recoveryService)

	return &Container{
		DB:              db,
		Redis:           rdb,
		AuthHandler:     authHandler,
		RoomHandler:     roomHandler,
		MessageHandler:  messageHandler,
		FileHandler:     fileHandler,
		UserHandler:     userHandler,
		RecoveryHandler: recoveryHandler,
	}

}

// ⭐ SetupRouter - 라우터 설정 메서드 추가
func (c *Container) SetupRouter() *gin.Engine {
	r := gin.Default()

	// 정적 파일
	c.setupStaticFiles(r)

	// 글로벌 미들웨어
	r.Use(middleware.Recovery())
	r.Use(middleware.ErrorHandler())
	r.Use(middleware.CORS())

	// Health Check
	r.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"status": "ok"})
	})

	// API v1
	api := r.Group("/api/v1")
	api.Use(middleware.APIRateLimit(c.Redis))

	// 인증 (Public)
	c.setupAuthRoutes(api)

	// 보호된 라우트
	c.setupProtectedRoutes(api)

	// 관리자
	c.setupAdminRoutes(api)

	return r
}

func (c *Container) setupStaticFiles(r *gin.Engine) {
	dir, _ := os.Getwd()
	r.Static("/uploads", filepath.Join(dir, "uploads"))
}

func (c *Container) setupAuthRoutes(api *gin.RouterGroup) {
	auth := api.Group("/auth")
	{
		auth.POST("/signup", c.AuthHandler.Signup)
		auth.POST("/login", c.AuthHandler.Login)
		auth.POST("/refresh", c.AuthHandler.Refresh)
	}
}

func (c *Container) setupProtectedRoutes(api *gin.RouterGroup) {
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware())
	{
		// 채팅방
		rooms := protected.Group("/rooms")
		{
			rooms.POST("", c.RoomHandler.CreateRoom)
			rooms.GET("", c.RoomHandler.GetMyRooms)
			rooms.GET("/:id", c.RoomHandler.GetRoom)
			rooms.POST("/:id/join", c.RoomHandler.JoinRoom)
			rooms.DELETE("/:id/leave", c.RoomHandler.LeaveRoom)
			rooms.POST("/:id/messages", c.MessageHandler.SendMessage)
			rooms.GET("/:id/messages", c.MessageHandler.GetMessages)
			rooms.POST("/:id/messages/image", c.MessageHandler.SendMessageWithImage)
		}

		// 파일
		files := protected.Group("/files")
		{
			files.GET("/:id", c.FileHandler.DownloadFile)
		}

		// 사용자
		users := protected.Group("/users")
		{
			users.GET("/recent", c.UserHandler.GetRecentUsers)
		}
	}
}

func (c *Container) setupAdminRoutes(api *gin.RouterGroup) {
	admin := api.Group("/admin")
	admin.Use(middleware.AuthMiddleware())
	{
		redis := admin.Group("/redis")
		{
			redis.POST("/recover", c.RecoveryHandler.RecoverFromDB)
			redis.GET("/health", c.RecoveryHandler.CheckHealth)
		}
	}
}

func migrateDB(db *gorm.DB) {
	err := db.AutoMigrate(
		&model.User{},
		&model.Room{},
		&model.RoomMember{},
		&model.Message{},
		&model.MessageRead{},
		&model.File{},
	)

	if err != nil {
		logger.Fatal("Database migration failed", zap.Error(err))
	}

	logger.Info("Database migration completed")

}
