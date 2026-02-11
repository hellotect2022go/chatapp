package model

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID               uint        `gorm:"primaryKey" json:"id"`
	RoomID           uuid.UUID   `gorm:"type:uuid;not null;index" json:"room_id"`
	UserUID          uuid.UUID   `gorm:"type:uuid;not null;index" json:"user_uid"`
	Content          string      `gorm:"type:text;not null" json:"content"`
	Type             string      `gorm:"type:varchar(20);not null" json:"type"`                      // text, image, file (메시지 콘텐츠 타입)
	MessageEventType MessageType `gorm:"type:varchar(20);not null;default:'chat'" json:"event_type"` // chat, join, leave (메시지 이벤트 타입)
	CreatedAt        time.Time   `gorm:"autoCreateTime;index" json:"created_at"`

	// 관계 Relationships N : 1
	User  User   `gorm:"foreignKey:UserUID;references:UID" json:"user,omitempty"` // ⭐ JSON에 포함
	Room  Room   `gorm:"foreignKey:RoomID;references:RoomID" json:"-"`
	Files []File `gorm:"many2many:message_files" json:"files,omitempty"`
}

type MessageRead struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	MessageID uint      `gorm:"not null;uniqueIndex:idx_message_user" json:"message_id"`
	UserUID   uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_message_user" json:"user_uid"`
	ReadAt    time.Time `gorm:"autoCreateTime" json:"read_at"`

	// 관계
	User User `gorm:"foreignKey:UserUID;references:UID" json:"-"`
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
	Type      MessageType `json:"type"` // chat, join, leave, typing, read (이벤트 타입)
	MessageID uint        `json:"message_id,omitempty"`
	Content   string      `json:"content,omitempty"`
	RoomID    string      `json:"room_id,omitempty"`  // ⭐ JSON에서는 string
	UserUID   string      `json:"user_uid,omitempty"` // ⭐ JSON에서는 string
	Nickname  string      `json:"nickname,omitempty"`
	Timestamp time.Time   `json:"timestamp"`

	// ⭐ 메시지 콘텐츠 타입
	ContentType string `json:"content_type,omitempty"` // "text", "image", "file"
	Files       []File `json:"files,omitempty"`        // 첨부 파일 정보
}
