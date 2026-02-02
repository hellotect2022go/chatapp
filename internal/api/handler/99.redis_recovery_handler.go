package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/hellotect2022go/chatapp/internal/api/service"
)

type RedisRecoveryHandler struct {
	recoveryService service.RedisRecoveryService
}

func NewRedisRecoveryHandler(recoveryService service.RedisRecoveryService) *RedisRecoveryHandler {
	return &RedisRecoveryHandler{
		recoveryService: recoveryService,
	}
}

// POST /api/v1/admin/redis/recover
func (h *RedisRecoveryHandler) RecoverFromDB(c *gin.Context) {
	// Redis 복원 실행
	if err := h.recoveryService.RecoverUserRoomFromDB(); err != nil {
		c.JSON(500, gin.H{
			"error":   "Redis recovery failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "Redis recovery completed successfully",
	})
}

// GET /api/v1/admin/redis/health
func (h *RedisRecoveryHandler) CheckHealth(c *gin.Context) {
	healthy, err := h.recoveryService.CheckRedisHealth()
	if err != nil {
		c.JSON(500, gin.H{
			"error":   "Health check failed",
			"message": err.Error(),
		})
		return
	}

	if !healthy {
		c.JSON(200, gin.H{
			"healthy": false,
			"message": "Redis may need recovery - no data found",
		})
		return
	}

	c.JSON(200, gin.H{
		"healthy": true,
		"message": "Redis is healthy",
	})
}
