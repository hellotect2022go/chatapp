package service

import (
	"strings"

	"github.com/google/uuid"
	"github.com/hellotect2022go/chatapp/internal/api/dto"
	"github.com/hellotect2022go/chatapp/internal/api/repository"
	"github.com/hellotect2022go/chatapp/internal/shared/errors"
	"github.com/hellotect2022go/chatapp/internal/shared/logger"
	"github.com/hellotect2022go/chatapp/internal/shared/model"
	"github.com/hellotect2022go/chatapp/internal/shared/util"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService interface {
	Signup(req *dto.SignupRequest) (*dto.AuthResponse, error)
	Login(email, password string) (*dto.AuthResponse, error)
	Refresh(refreshToken string) (string, error)
	Logout(uid uuid.UUID) error
}

type authService struct {
	userRepo    repository.UserRepository
	sessionRepo *repository.SessionRepository
}

func NewAuthService(
	userRepo repository.UserRepository,
	sessionRepo *repository.SessionRepository,
) AuthService {
	return &authService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
	}
}

func (s *authService) Signup(req *dto.SignupRequest) (*dto.AuthResponse, error) {
	logger.Info("User signup attempt",
		zap.String("email", req.Email),
		zap.String("nickname", req.Nickname),
	)

	// 1. 비밀번호 해싱
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.WrapInternal(err, "Failed to hash password")
	}

	// 2. 사용자 생성 (UID는 BeforeCreate에서 자동 생성)
	user := &model.User{
		Email:    req.Email,
		Password: string(hashedPassword),
		UserRole: "user",
		Nickname: req.Nickname,
		Age:      req.Age,
	}

	if err := s.userRepo.Create(user); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			logger.Warn("Signup failed: email already exists", zap.String("email", req.Email))
			return nil, errors.AlreadyExists("Email")
		}
		return nil, errors.WrapDatabase(err, "Failed to create user")
	}

	// 3. 토큰 생성 (UID 기반)
	token, err := util.GenerateToken(user.UID, user.Nickname, user.UserRole)
	if err != nil {
		return nil, errors.WrapInternal(err, "Failed to generate token")
	}

	user.RefreshToken = token.RefreshToken

	// 4. Refresh Token 저장
	if err := s.userRepo.UpdateRefreshToken(user.UID, token.RefreshToken); err != nil {
		return nil, errors.WrapDatabase(err, "Failed to save refresh token")
	}

	logger.Info("User signup successful",
		zap.String("uid", user.UID.String()),
		zap.String("email", user.Email),
		zap.String("nickname", user.Nickname),
	)

	return &dto.AuthResponse{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		User:         *user,
	}, nil
}

func (s *authService) Login(email, inputPassword string) (*dto.AuthResponse, error) {
	logger.Info("User login attempt", zap.String("email", email))

	// 1. 사용자 조회
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Warn("Login failed: user not found", zap.String("email", email))
			return nil, errors.Unauthorized("Invalid email or password")
		}
		return nil, errors.WrapDatabase(err, "Failed to find user")
	}

	// 2. 비밀번호 검증
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(inputPassword)); err != nil {
		logger.Warn("Login failed: invalid password",
			zap.String("email", email),
			zap.String("uid", user.UID.String()),
		)
		return nil, errors.Unauthorized("Invalid email or password")
	}

	// 3. JWT 생성 (UID 기반)
	token, err := util.GenerateToken(user.UID, user.Nickname, user.UserRole)
	if err != nil {
		return nil, errors.WrapInternal(err, "Failed to generate token")
	}

	// ⭐ 마지막 로그인 시간 업데이트
	if err := s.userRepo.UpdateLastLoginAt(user.UID); err != nil {
		logger.Warn("Failed to update last login time", zap.String("uid", user.UID.String()), zap.Error(err))
	}

	logger.Info("User login successful",
		zap.String("uid", user.UID.String()),
		zap.String("email", user.Email),
		zap.String("nickname", user.Nickname),
	)

	// 4. 응답
	return &dto.AuthResponse{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		User:         *user,
	}, nil
}

func (s *authService) Refresh(refreshToken string) (string, error) {
	// 1. JWT 검증
	claims, err := util.ValidateToken(refreshToken)
	if err != nil {
		return "", errors.New(errors.ErrInvalidToken, "Invalid or expired token")
	}

	// 2. UID string을 uuid.UUID로 파싱
	uid, err := uuid.Parse(claims.UserUID)
	if err != nil {
		return "", errors.BadRequest("Invalid UID format")
	}

	// 3. 새로운 Access Token 생성 (UID 기반)
	newToken, err := util.GenerateToken(uid, claims.Nickname, claims.UserRole)
	if err != nil {
		return "", errors.WrapInternal(err, "Failed to generate new token")
	}

	return newToken.AccessToken, nil
}

func (s *authService) Logout(uid uuid.UUID) error {
	// Refresh Token 삭제 (Redis)
	// SessionRepository를 UID 기반으로 수정 필요
	return nil
}
