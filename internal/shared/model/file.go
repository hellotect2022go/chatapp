package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// File - 모든 파일을 통합 관리하는 테이블
type File struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UID          uuid.UUID `gorm:"type:uuid;index" json:"uid"` // User.UID (업로드한 사용자)
	OriginalName string    `gorm:"type:varchar(255)" json:"original_name"`
	StoredName   string    `gorm:"type:varchar(255);not null" json:"stored_name"`
	FilePath     string    `gorm:"type:varchar(500);not null" json:"file_path"`
	FileURL      string    `gorm:"type:varchar(500);not null;index" json:"file_url"`
	MimeType     string    `gorm:"type:varchar(100)" json:"mime_type"`
	FileSize     int64     `gorm:"not null" json:"file_size"`
	Width        int       `json:"width"`
	Height       int       `json:"height"`
	ThumbnailURL string    `gorm:"type:varchar(500)" json:"thumbnail_url"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`

	//GORM은 모델에 DeletedAt 필드가 있으면 Delete 메서드를 호출했을 때 데이터를 실제로 삭제(Hard Delete)하지 않고,
	// **"이 데이터는 삭제됨"**이라고 표시(Soft Delete)만 합니다.
	//그래서 실제 SQL은 DELETE가 아닌 UPDATE 명령어로 deleted_at 컬럼에 현재 시간을 채우는 방식으로 동작
	// 데이터베이스에서 행(Row)을 완전히 제거하고 싶다면 Unscoped() 메서드를 사용
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 관계
	User User `gorm:"foreignKey:UID;references:UID" json:"-"`
}

// MessageFile - 메시지와 파일의 중간 테이블 (메시지 첨부파일용)
type MessageFile struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	MessageID uint      `gorm:"not null;index" json:"message_id"`
	FileID    uint      `gorm:"not null;index" json:"file_id"`
	Order     int       `gorm:"default:1" json:"order"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`

	// 관계
	Message Message `gorm:"foreignKey:MessageID" json:"-"`
	File    File    `gorm:"foreignKey:FileID" json:"file,omitempty"`
}
