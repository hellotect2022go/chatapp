package handler

import (
	"fmt"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hellotect2022go/chatapp/internal/api/dto"
	"github.com/hellotect2022go/chatapp/internal/api/service"
	"github.com/hellotect2022go/chatapp/internal/shared/model"
)

type MessageHandler struct {
	messageService service.MessageService
	fileService    service.FileService
}

func NewMessageHandler(messageService service.MessageService, fileService service.FileService) *MessageHandler {
	return &MessageHandler{
		messageService: messageService,
		fileService:    fileService,
	}
}

func (h *MessageHandler) SendMessage(c *gin.Context) {
	roomID := c.Param("id")
	userID, _ := c.Get("user_id")

	var id uint
	if _, err := fmt.Sscan(roomID, &id); err != nil {
		c.JSON(400, gin.H{"error": "Invalid room ID"})
		return
	}

	var req dto.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	message, err := h.messageService.SendMessage(id, userID.(uint), &req)
	if err != nil {
		if err.Error() == "you are not a member of this room" {
			c.JSON(403, gin.H{"error": err.Error()})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, message)
}

// GET /api/v1/rooms/:id/messages - 메시지 조회
func (h *MessageHandler) GetMessages(c *gin.Context) {
	roomID := c.Param("id")
	userID, _ := c.Get("user_id")

	var id uint
	fmt.Sscan(roomID, &id)

	// 쿼리 파라미터
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := h.messageService.GetMessages(id, userID.(uint), page, pageSize)
	if err != nil {
		c.JSON(403, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, result)
}

func (h *MessageHandler) SendMessageWithImage(c *gin.Context) {
	userID := c.GetUint("user_id")
	roomID, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	// 1. 이미지 파일 가져오기
	file, header, err := c.Request.FormFile("image")

	log.Println("image file header", header)

	if err != nil {
		log.Println("Failed to get image file", err)
		c.JSON(400, gin.H{"error": "No image file"})
		return
	}
	defer file.Close()

	// 2. Form Data에서 추가 정보 가져오기
	content := c.PostForm("content")

	// 3. 메시지 생성 (Type: "image") - Redis Pub/Sub 발행 없이
	message := &model.Message{
		RoomID:  uint(roomID),
		UserID:  userID,
		Content: content,
		Type:    "image",
	}

	if err := h.messageService.SaveMessage(message); err != nil {
		log.Println("Failed to send message", err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// 4. 이미지 업로드 및 File 레코드 생성
	fileModel, err := h.fileService.UploadImage(file, header, message.ID)
	if err != nil {
		log.Println("Failed to upload image", err)
		c.JSON(500, gin.H{"error": "Failed to upload image"})
		return
	}

	// ⭐ 5. 파일 정보를 포함하여 Redis Pub/Sub으로 발행
	message.Files = []model.File{*fileModel}
	if err := h.messageService.PublishMessage(message); err != nil {
		log.Println("Failed to publish message", err)
	}

	// 6. 응답
	c.JSON(200, gin.H{
		"message": message,
		"file":    fileModel,
	})
}
