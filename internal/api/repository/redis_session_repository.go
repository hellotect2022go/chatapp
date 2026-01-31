package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type SessionRepository struct {
	rdb *redis.Client
	ctx context.Context
}

func NewSessionRepository(rdb *redis.Client) *SessionRepository {
	return &SessionRepository{rdb: rdb, ctx: context.Background()}
}

func (r *SessionRepository) sessionKey(userID uint) string {
	return fmt.Sprintf("session:%d", userID)
}

// Refresh Token 저장
func (r *SessionRepository) Set(userID uint, refreshToken string, ttl time.Duration) error {
	key := r.sessionKey(userID)
	return r.rdb.Set(r.ctx, key, refreshToken, ttl).Err()
}

// Refresh Token 조회
func (r *SessionRepository) Get(userID uint) (string, error) {
	key := r.sessionKey(userID)
	return r.rdb.Get(r.ctx, key).Result()
}

// Refresh Token 삭제 (로그아웃)
func (r *SessionRepository) Delete(userID uint) error {
	key := r.sessionKey(userID)
	return r.rdb.Del(r.ctx, key).Err()
}

// Refresh Token 검증
func (r *SessionRepository) Validate(userID uint, refreshToken string) (bool, error) {
	stored, err := r.Get(userID)
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return stored == refreshToken, nil
}
