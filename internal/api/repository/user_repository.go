package repository

import (
	"context"

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

func (u *userRepository) userCacheKey(id uint) string {
	panic("not implemented") // TODO: Implement
}

func (u *userRepository) Create(user *model.User) error {
	panic("not implemented") // TODO: Implement
}

func (u *userRepository) FindByID(id uint) (*model.User, error) {
	panic("not implemented") // TODO: Implement
}

func (u *userRepository) FindByEmail(email string) (*model.User, error) {
	panic("not implemented") // TODO: Implement
}

func (u *userRepository) Update(user *model.User) error {
	panic("not implemented") // TODO: Implement
}

func (u *userRepository) Delete(id uint) error {
	panic("not implemented") // TODO: Implement
}

func (u *userRepository) UpdateRefreshToken(userID uint, token string) error {
	panic("not implemented") // TODO: Implement
}

func (u *userRepository) UpdateLastLoginAt(userID uint) error {
	panic("not implemented") // TODO: Implement
}

func (u *userRepository) FindRecentUsers(limit int) ([]*model.User, error) {
	panic("not implemented") // TODO: Implement
}

func (u *userRepository) FindAll() ([]*model.User, error) {
	panic("not implemented") // TODO: Implement
}
