package handler

import (
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	roomIDStr := c.Param("id")
	uidVal, exists := c.Get("uid")
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	// RoomID string → uuid.UUID
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid room ID"})
		return
	}

	// UID string → uuid.UUID
	userUID, err := uuid.Parse(uidVal.(string))
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid UID format"})
		return
	}

	var req dto.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	message, err := h.messageService.SendMessage(roomID, userUID, &req)
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
	roomIDStr := c.Param("id")
	uidVal, _ := c.Get("uid")

	// RoomID string → uuid.UUID
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid room ID"})
		return
	}

	// UID string → uuid.UUID
	userUID, err := uuid.Parse(uidVal.(string))
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid UID format"})
		return
	}

	// 쿼리 파라미터
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := h.messageService.GetMessages(roomID, userUID, page, pageSize)
	if err != nil {
		c.JSON(403, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, result)
}

func (h *MessageHandler) SendMessageWithImage(c *gin.Context) {
	uidVal, exists := c.Get("uid")
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	roomIDStr := c.Param("id")

	// RoomID string → uuid.UUID
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid room ID"})
		return
	}

	// UID string → uuid.UUID
	userUID, err := uuid.Parse(uidVal.(string))
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid UID format"})
		return
	}

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
		RoomID:  roomID,
		UserUID: userUID,
		Content: content,
		Type:    "image",
	}

	if err := h.messageService.SaveMessage(message); err != nil {
		log.Println("Failed to send message", err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// 4. 이미지 업로드 및 File 레코드 생성
	fileModel, err := h.fileService.UploadImage(file, header, userUID)
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
