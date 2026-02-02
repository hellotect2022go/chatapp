package dto

import "github.com/hellotect2022go/chatapp/internal/shared/model"

// SendMessageRequest - 메시지 전송 요청
type SendMessageRequest struct {
	Content string `json:"content" binding:"required,max=1000"`
	Type    string `json:"type" binding:"required,oneof=text image file"`
}

// MessageListResponse - 메시지 목록 응답 (페이지네이션)
type MessageListResponse struct {
	Messages   []*model.Message `json:"messages"`
	Total      int64            `json:"total"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
	TotalPages int              `json:"total_pages"`
}
