package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hellotect2022go/chatapp/internal/api/dto"
	"github.com/hellotect2022go/chatapp/internal/api/service"
	"github.com/hellotect2022go/chatapp/internal/shared/errors"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// ===== 기존 코드 (백업용 주석 처리) =====
/*
func (h *AuthHandler) Signup(c *gin.Context) {
	var req dto.SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errors.ValidationError(err.Error()))
		return
	}

	// Service 호출
	response, err := h.authService.Signup(&req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(201, response)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errors.ValidationError(err.Error()))
		return
	}

	// Service 호출
	response, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(200, response)
}
*/

// ⭐ Firebase 인증 가정: UID 기반 토큰 발급
func (h *AuthHandler) FirebaseAuth(c *gin.Context) {
	var req dto.FirebaseAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errors.ValidationError(err.Error()))
		return
	}

	// Service 호출: Firebase 인증된 사용자로 토큰 발급
	response, err := h.authService.FirebaseAuth(req.FirebaseUid, req.RefreshToken)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(200, response)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errors.ValidationError(err.Error()))
		return
	}

	// Service 호출
	accessToken, err := h.authService.Refresh(req.RefreshToken)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(200, gin.H{"access_token": accessToken})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	uidVal, exists := c.Get("uid")
	if !exists {
		c.Error(errors.Unauthorized("Authentication required"))
		return
	}

	// string → uuid.UUID 변환
	uid, err := uuid.Parse(uidVal.(string))
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid UID format"})
		return
	}

	// Service 호출
	if err := h.authService.Logout(uid); err != nil {
		c.Error(err)
		return
	}

	c.JSON(200, gin.H{"message": "Logged out successfully"})
}
