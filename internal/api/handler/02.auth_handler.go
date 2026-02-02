package handler

import (
	"github.com/gin-gonic/gin"
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
	userID, exists := c.Get("user_id")
	if !exists {
		c.Error(errors.Unauthorized("Authentication required"))
		return
	}

	// Service 호출
	if err := h.authService.Logout(userID.(uint)); err != nil {
		c.Error(err)
		return
	}

	c.JSON(200, gin.H{"message": "Logged out successfully"})
}
