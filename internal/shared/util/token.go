package util

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// 토큰 유효 기간 설정
const (
	AccessTokenDuration  = 15 * time.Hour     // 짧게 설정 (보안)
	RefreshTokenDuration = 7 * 24 * time.Hour // 길게 설정 (편의)
)

var jwtSecret = []byte("your-secret-key-change-this-in-production")

type Claims struct {
	UserUID  string `json:"user_uid"` // ⭐ JWT에서는 string으로 전송
	Nickname string `json:"nickname"`
	UserRole string `json:"user_role"`
	jwt.RegisteredClaims
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func GenerateToken(userUID uuid.UUID, nickname string, userRole string) (*TokenResponse, error) {

	claims := Claims{
		UserUID:  userUID.String(), // ⭐ UUID를 string으로 변환
		Nickname: nickname,
		UserRole: userRole,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(AccessTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "chatapp",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	accessToken, err := token.SignedString(jwtSecret)
	if err != nil {
		return nil, err
	}
	// refresh token 생성
	refreshTokenClaims := Claims{
		UserUID:  userUID.String(),
		UserRole: userRole,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(RefreshTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "chatapp",
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims)

	refreshTokenString, err := refreshToken.SignedString(jwtSecret)
	if err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenString,
	}, nil
}

func ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {

		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")

}
