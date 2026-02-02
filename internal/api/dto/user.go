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
