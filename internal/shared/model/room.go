package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RoomType string

const (
	RoomTypeDirect RoomType = "direct"
	RoomTypeGroup  RoomType = "group"
)

type Room struct {
	RoomID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"room_id"`
	Name      string    `gorm:"type:varchar(100)" json:"name"`
	Type      RoomType  `gorm:"type:varchar(20);not null" json:"type"`
	CreatedBy uuid.UUID `gorm:"type:uuid;not null" json:"created_by"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// 관계 Relationships
	Creator  User         `gorm:"foreignKey:CreatedBy;references:UID" json:"-"`
	Members  []RoomMember `gorm:"foreignKey:RoomID;references:RoomID" json:"-"`
	Messages []Message    `gorm:"foreignKey:RoomID;references:RoomID" json:"-"`
}

func (r *Room) BeforeCreate(tx *gorm.DB) error {
	if r.RoomID == uuid.Nil {
		r.RoomID, _ = uuid.NewV7()
	}
	return nil
}

type RoomMember struct {
	ID       uint      `gorm:"primaryKey" json:"id"`
	RoomID   uuid.UUID `gorm:"type:uuid;uniqueIndex:uk_room_user" json:"room_id"`
	UserUID  uuid.UUID `gorm:"type:uuid;uniqueIndex:uk_room_user" json:"user_uid"`
	JoinedAt time.Time `gorm:"autoCreateTime" json:"joined_at"`

	// 관계
	Room Room `gorm:"foreignKey:RoomID;references:RoomID" json:"-"`
	User User `gorm:"foreignKey:UserUID;references:UID" json:"-"`
}
