package handler

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hellotect2022go/chatapp/internal/api/service"
)

type PresenceHandler struct {
	presenceService *service.PresenceService
}

func NewPresenceHandler(presenceService *service.PresenceService) *PresenceHandler {
	return &PresenceHandler{presenceService: presenceService}
}

func (h *PresenceHandler) SetOnline(c *gin.Context) {
	uidVal, exists := c.Get("uid")
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}
	uid, err := uuid.Parse(uidVal.(string))
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid UID format"})
		return
	}

	if err := h.presenceService.SetOnline(uid); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Status updated"})
}

// POST /users/heartbeat
func (h *PresenceHandler) Heartbeat(c *gin.Context) {
	uidVal, exists := c.Get("uid")
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}
	uid, err := uuid.Parse(uidVal.(string))
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid UID format"})
		return
	}

	if err := h.presenceService.Heartbeat(uid); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Heartbeat received"})
}

// GET /users/:id/online
func (h *PresenceHandler) IsOnline(c *gin.Context) {
	uidStr := c.Param("id")
	uid, err := uuid.Parse(uidStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid UID format"})
		return
	}

	online, err := h.presenceService.IsOnline(uid)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"online": online})
}

// POST /users/online-status (여러 사용자)
func (h *PresenceHandler) GetOnlineStatus(c *gin.Context) {
	var req struct {
		UserUIDs []string `json:"user_uids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// string 배열을 uuid.UUID 배열로 변환
	userUIDs := make([]uuid.UUID, 0, len(req.UserUIDs))
	for _, uidStr := range req.UserUIDs {
		uid, err := uuid.Parse(uidStr)
		if err != nil {
			c.JSON(400, gin.H{"error": fmt.Sprintf("Invalid UID: %s", uidStr)})
			return
		}
		userUIDs = append(userUIDs, uid)
	}

	status, err := h.presenceService.GetOnlineStatus(userUIDs)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// uuid.UUID 키를 string으로 변환해서 반환
	statusMap := make(map[string]bool)
	for uid, online := range status {
		statusMap[uid.String()] = online
	}

	c.JSON(200, statusMap)
}
