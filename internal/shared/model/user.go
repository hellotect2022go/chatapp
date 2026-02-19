package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User - 사용자 정보 (UID가 PK)
type User struct {
	UID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"uid"`
	FirebaseUid  string    `gorm:"type:varchar(30)" json:"firebase_uid"`
	Email        string    `gorm:"type:varchar(255);uniqueIndex" json:"email,omitempty"`
	Password     string    `gorm:"type:varchar(255)" json:"-"`
	RefreshToken string    `gorm:"type:varchar(512)" json:"-"`
	UserRole     string    `gorm:"type:varchar(20);default:'user'" json:"user_role"`
	Nickname     string    `gorm:"type:varchar(50)" json:"nickname"`
	Age          int       `gorm:"type:int" json:"age"`
	Gender       string    `gorm:"type:varchar(10)" json:"gender"`
	Region       string    `gorm:"type:varchar(50)" json:"region"`
	Bio          string    `gorm:"type:text" json:"bio"`
	Avatar       string    `gorm:"type:varchar(255)" json:"avatar"`
	LastLoginAt  time.Time `gorm:"type:timestamp" json:"last_login_at,omitempty"`
	CreatedAt    time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"`
	IsCompleted  bool      `gorm:"default:false" json:"isCompleted"`

	// Relationships - constraint 제약조건은 제거하고 참조만 유지
	ProfileImages []ProfileImage `gorm:"foreignKey:UID;references:UID;constraint:OnDelete:CASCADE" json:"profile_images,omitempty"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.UID == uuid.Nil {
		u.UID, _ = uuid.NewV7()
	}
	return nil
}

// ProfileImage - 사용자 프로필 이미지 (1:N)
type ProfileImage struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UID       uuid.UUID `gorm:"type:uuid;index;not null" json:"uid"`
	FileID    uint      `gorm:"not null" json:"file_id"`
	Order     int       `gorm:"default:0" json:"order"`
	IsPrimary bool      `gorm:"default:false" json:"is_primary"`
	CreatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`

	// Relationships
	User User `gorm:"foreignKey:UID;references:UID" json:"-"`
	File File `gorm:"foreignKey:FileID" json:"file,omitempty"`
}
