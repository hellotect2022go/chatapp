package model

import "time"

type RoomType string

const (
	RoomTypeDirect RoomType = "direct"
	RoomTypeGroup  RoomType = "group"
)

type Room struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"type:varchar(100)" json:"name"`
	Type      RoomType  `gorm:"type:varchar(20);not null" json:"type"`
	CreatedBy uint      `gorm:"not null" json:"created_by"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// 관계 Relationships
	Members  []RoomMember `gorm:"foreignKey:RoomID" json:"-"`
	Messages []Message    `gorm:"foreignKey:RoomID" json:"-"`
}

type RoomMember struct {
	ID uint `gorm:"primaryKey" json:"id"`
	// Unique 속성을 강조할 때 (추천)
	RoomID   uint      `gorm:"uniqueIndex:uk_room_user" json:"room_id"`
	UserID   uint      `gorm:"uniqueIndex:uk_room_user" json:"user_id"`
	JoinedAt time.Time `gorm:"autoCreateTime" json:"joined_at"`
}
