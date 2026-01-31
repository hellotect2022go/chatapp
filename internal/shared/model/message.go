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
