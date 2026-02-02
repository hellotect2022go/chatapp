package model

import "time"

type Message struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	RoomID    uint      `gorm:"not null;index" json:"room_id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	Type      string    `gorm:"type:varchar(20);not null" json:"type"` // text, image, file
	CreatedAt time.Time `gorm:"autoCreateTime;index" json:"created_at"`

	// 관계 Relationships N : 1
	User  User   `gorm:"foreignKey:UserID" json:"user,omitempty"` // ⭐ JSON에 포함
	Room  Room   `gorm:"foreignKey:RoomID" json:"-"`
	Files []File `gorm:"foreignKey:MessageID" json:"files,omitempty"`
}

type MessageRead struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	MessageID uint      `gorm:"not null;uniqueIndex:idx_message_user" json:"message_id"`
	UserID    uint      `gorm:"not null;uniqueIndex:idx_message_user" json:"user_id"`
	ReadAt    time.Time `gorm:"autoCreateTime" json:"read_at"`
}

type MessageType string

const (
	MessageTypeChat   MessageType = "chat"   // 채팅 메세지
	MessageTypeJoin   MessageType = "join"   // 입장 메세지
	MessageTypeLeave  MessageType = "leave"  // 퇴장 메세지
	MessageTypeTyping MessageType = "typing" // 입력 중 메세지
	MessageTypeRead   MessageType = "read"   // 읽음 메세지

	// MessageTypeImage    MessageType = "image"    // 이미지 메세지
	// MessageTypeVoice    MessageType = "voice"    // 음성 메세지
	// MessageTypeFile     MessageType = "file"     // 파일 메세지
	// MessageTypeLocation MessageType = "location" // 위치 메세지
	// MessageTypeSystem   MessageType = "system"   // 시스템 메세지
)

type WSMessage struct {
	Type      MessageType `json:"type"`
	MessageID uint        `json:"message_id,omitempty"`
	Content   string      `json:"content,omitempty"`
	RoomID    uint        `json:"room_id,omitempty"`
	UserID    uint        `json:"user_id,omitempty"`
	Nickname  string      `json:"nickname,omitempty"`
	Timestamp time.Time   `json:"timestamp"`

	// ⭐ 추가: 이미지 메시지 지원
	MessageType string `json:"message_type,omitempty"` // "text", "image", "file"
	Files       []File `json:"files,omitempty"`        // 첨부 파일 정보
}
