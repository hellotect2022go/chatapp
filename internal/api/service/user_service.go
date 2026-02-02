package service

import (
	"github.com/hellotect2022go/chatapp/internal/api/repository"
	"github.com/hellotect2022go/chatapp/internal/shared/errors"
	"github.com/hellotect2022go/chatapp/internal/shared/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService interface {
	GetByID(id uint) (*model.User, error)
	Update(id uint, nickname string, age int) error
	Delete(id uint) error
	ChangePassword(id uint, newPassword string) error
	GetAll() ([]*model.User, error)
	GetRecentUsers(limit int) ([]*model.User, error)
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

func (s *userService) GetByID(id uint) (*model.User, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("User")
		}
		return nil, errors.WrapDatabase(err, "Failed to find user")
	}
	return user, nil
}

func (s *userService) Update(id uint, nickname string, age int) error {
	// 1. 사용자 조회
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NotFound("User")
		}
		return errors.WrapDatabase(err, "Failed to find user")
	}

	// 2. 사용자 업데이트
	if nickname != "" {
		user.Nickname = nickname
	}

	if age > 0 {
		user.Age = age
	}

	if err := s.userRepo.Update(user); err != nil {
		return errors.WrapDatabase(err, "Failed to update user")
	}
	return nil
}

func (s *userService) Delete(id uint) error {
	if err := s.userRepo.Delete(id); err != nil {
		return errors.WrapDatabase(err, "Failed to delete user")
	}
	return nil
}

func (s *userService) ChangePassword(id uint, newPassword string) error {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NotFound("User")
		}
		return errors.WrapDatabase(err, "Failed to find user")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.WrapInternal(err, "Failed to hash password")
	}

	user.Password = string(hashedPassword)

	if err := s.userRepo.Update(user); err != nil {
		return errors.WrapDatabase(err, "Failed to update password")
	}
	return nil
}

func (s *userService) GetAll() ([]*model.User, error) {
	users, err := s.userRepo.FindAll()
	if err != nil {
		return nil, errors.WrapDatabase(err, "Failed to load users")
	}
	return users, nil
}

// GetRecentUsers - 최근 로그인한 사용자 목록 조회
func (s *userService) GetRecentUsers(limit int) ([]*model.User, error) {
	users, err := s.userRepo.FindRecentUsers(limit)
	if err != nil {
		return nil, errors.WrapDatabase(err, "Failed to load recent users")
	}
	return users, nil
}
