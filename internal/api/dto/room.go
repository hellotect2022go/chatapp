package dto

import "github.com/hellotect2022go/chatapp/internal/shared/model"

// CreateRoomRequest - 채팅방 생성 요청
type CreateRoomRequest struct {
	Name    string `json:"name" binding:"required,min=1,max=100"`
	Type    string `json:"type" binding:"required,oneof=direct group"`
	Members []uint `json:"members" binding:"required,min=1"` // 초대할 사용자 ID들
}

// RoomDetailResponse - 채팅방 상세 응답 (멤버 정보 포함)
type RoomDetailResponse struct {
	Room    *model.Room   `json:"room"`
	Members []*model.User `json:"members"`
}
