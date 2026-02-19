package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/hellotect2022go/chatapp/internal/api/dto"
	"github.com/hellotect2022go/chatapp/internal/shared/model"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type UserRepository interface {
	userCacheKey(uid uuid.UUID) string
	Create(user *model.User) error
	FindByUID(uid uuid.UUID) (*model.User, error)
	FindByEmail(email string) (*model.User, error)
	Update(user *model.User) error
	Delete(uid uuid.UUID) error
	UpdateRefreshToken(uid uuid.UUID, token string) error
	UpdateLastLoginAt(uid uuid.UUID) error
	FindRecentUsers(limit int) ([]*model.User, error)
	FindAll() ([]*model.User, error)
	FindUsersWithFilter(filter dto.UserListFilter) ([]*model.User, int64, error)
	DeleteProfileImagesByURLs(uid uuid.UUID, imageUrls []string) error // ⭐ 추가

	DB() *gorm.DB
}

type userRepository struct {
	db  *gorm.DB
	rdb *redis.Client
	ctx context.Context
}

func NewUserRepository(db *gorm.DB, rdb *redis.Client) UserRepository {
	return &userRepository{db, rdb, context.Background()}
}

func (r *userRepository) userCacheKey(uid uuid.UUID) string {
	return fmt.Sprintf("user:%s", uid.String())
}

func (r *userRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) FindByUID(uid uuid.UUID) (*model.User, error) {
	// 1. Redis 캐시 조회
	cacheKey := r.userCacheKey(uid)
	cached, err := r.rdb.Get(r.ctx, cacheKey).Result()
	if err == nil {
		var user model.User
		if err := json.Unmarshal([]byte(cached), &user); err == nil {
			return &user, nil
		}
	}

	// 2. DB 조회
	var user model.User
	result := r.db.
		Preload("ProfileImages.File").
		Where("uid = ?", uid).
		First(&user)
	if result.Error != nil {
		return nil, result.Error
	}

	// 3. Redis 캐시 저장 (10분 TTL)
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

	// 2. Redis 캐시 무효화
	cacheKey := r.userCacheKey(user.UID)
	r.rdb.Del(r.ctx, cacheKey)

	return nil
}

func (r *userRepository) Delete(uid uuid.UUID) error {
	// 1. DB 삭제
	if err := r.db.Where("uid = ?", uid).Delete(&model.User{}).Error; err != nil {
		return err
	}

	// 2. Redis 캐시 무효화
	cacheKey := r.userCacheKey(uid)
	r.rdb.Del(r.ctx, cacheKey)

	return nil
}

func (r *userRepository) UpdateRefreshToken(uid uuid.UUID, token string) error {
	return r.db.Model(&model.User{}).
		Where("uid = ?", uid).
		Update("refresh_token", token).Error
}

func (r *userRepository) UpdateLastLoginAt(uid uuid.UUID) error {
	return r.db.Model(&model.User{}).
		Where("uid = ?", uid).
		Update("last_login_at", time.Now()).Error
}

func (r *userRepository) FindRecentUsers(limit int) ([]*model.User, error) {
	var users []*model.User
	err := r.db.
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

// FindUsersWithFilter - 필터링된 사용자 목록 조회
func (r *userRepository) FindUsersWithFilter(filter dto.UserListFilter) ([]*model.User, int64, error) {
	// User 테이블 기반으로 쿼리
	query := r.db.Model(&model.User{})

	// 필터 적용
	if filter.Gender != "" {
		query = query.Where("gender = ?", filter.Gender)
	}
	if filter.Region != "" {
		query = query.Where("region = ?", filter.Region)
	}
	if filter.MinAge > 0 {
		query = query.Where("age >= ?", filter.MinAge)
	}
	if filter.MaxAge > 0 {
		query = query.Where("age <= ?", filter.MaxAge)
	}
	if filter.Search != "" {
		query = query.Where("nickname LIKE ?", "%"+filter.Search+"%")
	}

	// 전체 개수 조회
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 페이징
	page := filter.Page
	if page <= 0 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	// User 목록 조회
	var users []*model.User
	if err := query.
		Preload("ProfileImages.File").
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// DeleteProfileImagesByURLs - 특정 이미지 URL들을 DB에서 삭제
func (r *userRepository) DeleteProfileImagesByURLs(uid uuid.UUID, imageUrls []string) error {
	if len(imageUrls) == 0 {
		return nil
	}

	// 1. URL로 File ID 조회
	var files []model.File
	if err := r.db.Where("file_url IN ?", imageUrls).Delete(&files).Error; err != nil {
		return err
	}

	log.Printf("✅ ProfileImage 삭제 완료: %d개 (UID: %s)", len(imageUrls), uid)

	return nil
}

// DB - DB 인스턴스 반환 (service에서 사용)
func (r *userRepository) DB() *gorm.DB {
	return r.db
}
