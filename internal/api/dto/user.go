package dto

import "github.com/hellotect2022go/chatapp/internal/shared/model"

type SignupRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Nickname string `json:"nickname" binding:"required,min=2,max=20"`
	Age      int    `json:"age" binding:"required,gte=0,lte=150"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type UpdateMeRequest struct {
	Nickname string `json:"nickname" binding:"omitempty,min=2,max=20"`
	Age      int    `json:"age" binding:"omitempty,gte=0,lte=150"`
}

type ChangePasswordRequest struct {
	Password string `json:"password" binding:"required,min=8"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenResponse struct {
	AccessToken string `json:"access_token" binding:"required"`
}

type AuthResponse struct {
	AccessToken  string     `json:"access_token"`
	RefreshToken string     `json:"refresh_token"`
	User         model.User `json:"user"`
}

// 프로필 생성 요청
type CreateProfileRequest struct {
	UID          string `json:"uid" binding:"required"` // Firebase UID
	RefreshToken string `json:"refreshToken"`           // Firebase RefreshToken
	Nickname     string `json:"nickname" binding:"required,min=2,max=20"`
	Age          int    `json:"age,string" binding:"required,gte=19,lte=150"`
	Gender       string `json:"gender" binding:"required,oneof=male female other"`
	Region       string `json:"region" binding:"required"`
	Bio          string `json:"bio" binding:"max=500"`
	Avatar       string `json:"avatar"`
	ProfileImage string `json:"profileImage"` // 이미 업로드된 이미지 URL
}

// 프로필 수정 요청
type UpdateProfileRequest struct {
	Nickname     string `json:"nickname" binding:"omitempty,min=2,max=20"`
	Age          int    `json:"age" binding:"omitempty,gte=19,lte=150"`
	Gender       string `json:"gender" binding:"omitempty,oneof=male female other"`
	Region       string `json:"region"`
	Bio          string `json:"bio" binding:"max=500"`
	Avatar       string `json:"avatar"`
	ProfileImage string `json:"profileImage"`
}

// 사용자 목록 필터
type UserListFilter struct {
	Gender   string `form:"gender"`
	Region   string `form:"region"`
	MinAge   int    `form:"minAge"`
	MaxAge   int    `form:"maxAge"`
	Search   string `form:"search"`
	Page     int    `form:"page"`
	PageSize int    `form:"pageSize"`
}

// 사용자 목록 응답
type UserListResponse struct {
	Users      []*model.User `json:"users"`
	Total      int64         `json:"total"`
	Page       int           `json:"page"`
	PageSize   int           `json:"pageSize"`
	TotalPages int           `json:"totalPages"`
}
