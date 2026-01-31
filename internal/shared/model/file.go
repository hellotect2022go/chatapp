package model

import (
	"time"

	"gorm.io/gorm"
)

type File struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	MessageID    uint           `gorm:"not null;index" json:"message_id"`
	OriginalName string         `gorm:"type:varchar(255)" json:"original_name"`
	StoredName   string         `gorm:"type:varchar(255);not null" json:"stored_name"`
	FilePath     string         `gorm:"type:varchar(500);not null" json:"file_path"`
	FileURL      string         `gorm:"type:varchar(500);not null" json:"file_url"`
	MimeType     string         `gorm:"type:varchar(100)" json:"mime_type"`
	FileSize     int64          `gorm:"not null" json:"file_size"`
	Width        int            `json:"width"`
	Height       int            `json:"height"`
	ThumbnailURL string         `gorm:"type:varchar(500)" json:"thumbnail_url"`
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"created_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// 관계
	Message Message `gorm:"foreignKey:MessageID" json:"-"`
}
