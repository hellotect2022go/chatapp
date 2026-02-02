package handler

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/hellotect2022go/chatapp/internal/api/service"
)

type PresenceHandler struct {
	presenceService *service.PresenceService
}

func NewPresenceHandler(presenceService *service.PresenceService) *PresenceHandler {
	return &PresenceHandler{presenceService: presenceService}
}

func (h *PresenceHandler) SetOnline(c *gin.Context) {
	userID, _ := c.Get("user_id")

	if err := h.presenceService.SetOnline(userID.(uint)); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Status updated"})
}

// POST /users/heartbeat
func (h *PresenceHandler) Heartbeat(c *gin.Context) {
	userID, _ := c.Get("user_id")

	if err := h.presenceService.Heartbeat(userID.(uint)); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Heartbeat received"})
}

// GET /users/:id/online
func (h *PresenceHandler) IsOnline(c *gin.Context) {
	id := c.Param("id")
	var userID uint
	fmt.Sscan(id, &userID)

	online, err := h.presenceService.IsOnline(userID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"online": online})
}

// POST /users/online-status (여러 사용자)
func (h *PresenceHandler) GetOnlineStatus(c *gin.Context) {
	var req struct {
		UserIDs []uint `json:"user_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	status, err := h.presenceService.GetOnlineStatus(req.UserIDs)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, status)
}
