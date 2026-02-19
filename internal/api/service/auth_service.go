package service

import (
	"github.com/google/uuid"
	"github.com/hellotect2022go/chatapp/internal/api/dto"
	"github.com/hellotect2022go/chatapp/internal/api/repository"
	"github.com/hellotect2022go/chatapp/internal/shared/errors"
	"github.com/hellotect2022go/chatapp/internal/shared/logger"
	"github.com/hellotect2022go/chatapp/internal/shared/model"
	"github.com/hellotect2022go/chatapp/internal/shared/util"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AuthService interface {
	// ===== 기존 메서드 (백업용 주석) =====
	// Signup(req *dto.SignupRequest) (*dto.AuthResponse, error)
	// Login(email, password string) (*dto.AuthResponse, error)

	// ⭐ Firebase 인증 가정
	FirebaseAuth(firebase_uid, firebaseRefreshToken string) (*dto.AuthResponse, error)
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

// ===== 기존 코드 (백업용 주석 처리) =====
/*
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
*/

// ⭐ FirebaseAuth - Firebase 인증된 사용자로 토큰 발급
func (s *authService) FirebaseAuth(firebase_uid, firebaseRefreshToken string) (*dto.AuthResponse, error) {
	logger.Info("Firebase auth attempt", zap.String("firebase_uid", firebase_uid))

	//1. 사용자 조회 (UID 기반)
	user, err := s.userRepo.FindByFirebaseUID(firebase_uid)

	// CASE A : 신규사용자 (DB 에 기록 없음)
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return nil, errors.WrapDatabase(err, "Failed to find user")
		}

		newUser := &model.User{
			FirebaseUid: firebase_uid,
			IsCompleted: false,
		}

		if err := s.userRepo.Create(newUser); err != nil {
			return nil, errors.WrapDatabase(err, "Failed to create user")
		}
		// 새 유저 정보만 반환 (프로필 미완료 → 클라이언트는 프로필 등록 화면)
		return &dto.AuthResponse{
			User:            *newUser,
			ProfileComplete: false,
		}, nil
	}

	// CASE B : 기존 사용자지만 프로필 미완성 (닉네임 없음) → 프로필 등록 화면
	if !user.IsCompleted {
		return &dto.AuthResponse{
			User:            *user,
			ProfileComplete: false,
		}, nil
	}

	// 닉네임 있으면 완료로 간주 (레거시 또는 IsCompleted 미반영 보정)
	if !user.IsCompleted && user.Nickname != "" {
		user.IsCompleted = true
		_ = s.userRepo.Update(user)
	}

	// CASE C : 프로필 완성된 사용자 -> 토큰 발급
	token, err := util.GenerateToken(user.UID, user.Nickname, user.UserRole)
	if err != nil {
		return nil, errors.WrapInternal(err, "Failed to generate token")
	}

	// DB 에 RefreshToken 저장 (필요시)
	user.RefreshToken = token.RefreshToken
	s.userRepo.Update(user)

	return &dto.AuthResponse{
		User:            *user,
		AccessToken:     token.AccessToken,
		RefreshToken:    token.RefreshToken,
		ProfileComplete: true, // 클라이언트는 메인 화면으로 이동
	}, nil

}

// ⭐ FirebaseAuth - Firebase 인증된 사용자로 토큰 발급
// func (s *authService) FirebaseAuth(uid uuid.UUID, firebaseRefreshToken string) (*dto.AuthResponse, error) {
// 	logger.Info("Firebase auth attempt", zap.String("uid", uid.String()))

// 	//1. 사용자 조회 (UID 기반)
// 	user, err := s.userRepo.FindByUID(uid)
// 	if err != nil {
// 		if err == gorm.ErrRecordNotFound {
// 			logger.Warn("Firebase auth failed: user not found", zap.String("uid", uid.String()))
// 			return nil, errors.NotFound("User not found. Please create profile first.")
// 		}
// 		return nil, errors.WrapDatabase(err, "Failed to find user")
// 	}

// 	//2. JWT 생성 (UID 기반)
// 	token, err := util.GenerateToken(user.UID, user.Nickname, user.UserRole)
// 	if err != nil {
// 		return nil, errors.WrapInternal(err, "Failed to generate token")
// 	}

// 	//4. 마지막 로그인 시간 업데이트
// 	if err := s.userRepo.UpdateLastLoginAt(user.UID); err != nil {
// 		logger.Warn("Failed to update last login time", zap.String("uid", user.UID.String()), zap.Error(err))
// 	}

// 	logger.Info("Firebase auth successful",
// 		zap.String("uid", user.UID.String()),
// 		zap.String("nickname", user.Nickname),
// 	)

// 	// 5. 응답
// 	return &dto.AuthResponse{
// 		AccessToken:  token.AccessToken,
// 		RefreshToken: token.RefreshToken,
// 		User:         *user,
// 	}, nil
// }

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
