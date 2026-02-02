package handler

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/hellotect2022go/chatapp/internal/api/dto"
	"github.com/hellotect2022go/chatapp/internal/api/service"
	"github.com/hellotect2022go/chatapp/internal/shared/errors"
)

type RoomHandler struct {
	roomService    service.RoomService
	messageService service.MessageService
}

func NewRoomHandler(roomService service.RoomService, messageService service.MessageService) *RoomHandler {
	return &RoomHandler{
		roomService:    roomService,
		messageService: messageService,
	}
}

func (h *RoomHandler) CreateRoom(c *gin.Context) {
	var req dto.CreateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errors.ValidationError(err.Error()))
		return
	}

	userID, _ := c.Get("user_id")

	room, err := h.roomService.CreateRoom(&req, userID.(uint))
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(201, room)
}

// GET /api/v1/rooms
func (h *RoomHandler) GetMyRooms(c *gin.Context) {
	userID, _ := c.Get("user_id")

	rooms, err := h.roomService.GetMyRooms(userID.(uint))
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(200, rooms)
}

// GET /api/v1/rooms/:id
func (h *RoomHandler) GetRoom(c *gin.Context) {
	roomID := c.Param("id")
	userID, _ := c.Get("user_id")

	var id uint
	_, err := fmt.Sscan(roomID, &id)
	if err != nil {
		c.Error(errors.BadRequest("Invalid room ID"))
		return
	}

	room, err := h.roomService.GetRoom(id, userID.(uint))
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(200, room)
}

// POST /api/v1/rooms/:id/join - ⭐ 채팅방 입장 API (실시간 알림만, 메시지 저장 안 함)
// 이미 멤버인 사용자가 방을 선택할 때 호출 (선택 사항)
func (h *RoomHandler) JoinRoom(c *gin.Context) {
	roomID := c.Param("id")
	userID, _ := c.Get("user_id")

	var id uint
	_, err := fmt.Sscan(roomID, &id)
	if err != nil {
		c.Error(errors.BadRequest("Invalid room ID"))
		return
	}

	// ⭐ WebSocket 실시간 알림만 (메시지 저장 안 함)
	if err := h.roomService.JoinRoom(id, userID.(uint)); err != nil {
		c.Error(err)
		return
	}

	c.JSON(200, gin.H{"message": "Joined room successfully"})
}

// DELETE /api/v1/rooms/:id/leave - ⭐ 채팅방 퇴장
func (h *RoomHandler) LeaveRoom(c *gin.Context) {
	roomID := c.Param("id")
	userID, _ := c.Get("user_id")

	var id uint
	_, err := fmt.Sscan(roomID, &id)
	if err != nil {
		c.Error(errors.BadRequest("Invalid room ID"))
		return
	}

	// ⭐ 퇴장 처리 (DB에 퇴장 메시지 저장 + WebSocket 알림)
	if err := h.roomService.LeaveRoom(id, userID.(uint)); err != nil {
		c.Error(err)
		return
	}

	c.Status(204)
}
