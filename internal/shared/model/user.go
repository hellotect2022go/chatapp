package model

import "time"

type User struct {
	ID           uint       `json:"id" gorm:"primaryKey"`
	Email        string     `json:"email" gorm:"type:varchar(100);not null;unique"`
	Password     string     `json:"-" gorm:"type:varchar(100);not null"`
	Nickname     string     `json:"nickname" gorm:"type:varchar(100);not null"`
	UserRole     string     `json:"user_role" gorm:"type:varchar(100);not null"`
	Age          int        `json:"age" gorm:"not null"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty" gorm:"index"` // ⭐ 추가: 최근 로그인 시간
	CreatedAt    time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	RefreshToken string     `json:"-" gorm:"type:text"` // DB에 저장할 리프레시 토큰
}
