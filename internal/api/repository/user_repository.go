package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/hellotect2022go/chatapp/internal/shared/model"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type UserRepository interface {
	userCacheKey(id uint) string
	Create(user *model.User) error
	FindByID(id uint) (*model.User, error)
	FindByEmail(email string) (*model.User, error)
	Update(user *model.User) error
	Delete(id uint) error
	UpdateRefreshToken(userID uint, token string) error
	UpdateLastLoginAt(userID uint) error              // ⭐ 추가: 로그인 시간 업데이트
	FindRecentUsers(limit int) ([]*model.User, error) // ⭐ 추가: 최근 로그인 사용자 조회
	FindAll() ([]*model.User, error)
}

type userRepository struct {
	db  *gorm.DB
	rdb *redis.Client
	ctx context.Context
}

func NewUserRepository(db *gorm.DB, rdb *redis.Client) UserRepository {
	return &userRepository{db, rdb, context.Background()}
}

func (r *userRepository) userCacheKey(id uint) string {
	return fmt.Sprintf("user:%d", id)
}

func (r *userRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) FindByID(id uint) (*model.User, error) {
	// 1. Redis 에서 조회 (cache-aside)
	cacheKey := r.userCacheKey(id)
	cached, err := r.rdb.Get(r.ctx, cacheKey).Result()
	if err == nil {
		var user model.User
		if err := json.Unmarshal([]byte(cached), &user); err == nil {
			return &user, nil
		}
	}

	// 2. Redis 에서 데이터가 없으면 DB 에서 조회
	var user model.User
	result := r.db.First(&user, id)
	if result.Error != nil {
		return nil, result.Error
	}

	// 3. Redis 에 저장 (10분 TTL)
	data, _ := json.Marshal(user)
	r.rdb.Set(r.ctx, cacheKey, data, 10*time.Minute)

	return &user, nil
}

func (r *userRepository) FindByEmail(email string) (*model.User, error) {
	var user model.User
	result := r.db.Where("email = ?", email).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func (r *userRepository) Update(user *model.User) error {
	// 1. DB 업데이트
	if err := r.db.Save(user).Error; err != nil {
		return err
	}

	// 2. 캐시 무효화 (cache invalidation)
	cacheKey := r.userCacheKey(user.ID)
	r.rdb.Del(r.ctx, cacheKey)

	return nil
}

func (r *userRepository) Delete(id uint) error {
	// 1. DB 삭제
	result := r.db.Delete(&model.User{}, id)
	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}

	if result.Error != nil {
		return result.Error
	}

	// 2. 캐시 무효화 (cache invalidation)
	cacheKey := r.userCacheKey(id)
	r.rdb.Del(r.ctx, cacheKey)
	return nil
}

func (r *userRepository) UpdateRefreshToken(userID uint, token string) error {
	return r.db.Model(&model.User{}).Where("id = ?", userID).Update("refresh_token", token).Error
}

func (r *userRepository) UpdateLastLoginAt(userID uint) error {
	now := time.Now()
	if err := r.db.Model(&model.User{}).Where("id = ?", userID).Update("last_login_at", now).Error; err != nil {
		return err
	}

	// 캐시 무효화
	cacheKey := r.userCacheKey(userID)
	r.rdb.Del(r.ctx, cacheKey)

	return nil
}

func (r *userRepository) FindRecentUsers(limit int) ([]*model.User, error) {
	var users []*model.User
	err := r.db.
		Where("last_login_at IS NOT NULL").
		Order("last_login_at DESC").
		Limit(limit).
		Find(&users).Error

	if err != nil {
		return nil, err
	}

	return users, nil
}

func (r *userRepository) FindAll() ([]*model.User, error) {
	var users []*model.User
	result := r.db.Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}
