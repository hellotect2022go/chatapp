package service

import (
	"github.com/google/uuid"
	"github.com/hellotect2022go/chatapp/internal/api/dto"
	"github.com/hellotect2022go/chatapp/internal/api/repository"
	"github.com/hellotect2022go/chatapp/internal/shared/errors"
	"github.com/hellotect2022go/chatapp/internal/shared/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService interface {
	GetByUID(uid uuid.UUID) (*model.User, error)
	Update(uid uuid.UUID, nickname string, age int) error
	Delete(uid uuid.UUID) error
	ChangePassword(uid uuid.UUID, newPassword string) error
	GetAll() ([]*model.User, error)
	GetRecentUsers(limit int) ([]*model.User, error)
	CreateProfileByUID(req dto.CreateProfileRequest) error
	UpdateProfile(uid uuid.UUID, req dto.UpdateProfileRequest) error
	GetUsersWithFilter(filter dto.UserListFilter) (*dto.UserListResponse, error)
	AddProfileImage(uid uuid.UUID, fileID uint, isPrimary bool) error
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

func (s *userService) GetByUID(uid uuid.UUID) (*model.User, error) {
	user, err := s.userRepo.FindByUID(uid)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("User")
		}
		return nil, errors.WrapDatabase(err, "Failed to find user")
	}
	return user, nil
}

func (s *userService) Update(uid uuid.UUID, nickname string, age int) error {
	// 1. 사용자 조회
	user, err := s.userRepo.FindByUID(uid)
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

func (s *userService) Delete(uid uuid.UUID) error {
	if err := s.userRepo.Delete(uid); err != nil {
		return errors.WrapDatabase(err, "Failed to delete user")
	}
	return nil
}

func (s *userService) ChangePassword(uid uuid.UUID, newPassword string) error {
	user, err := s.userRepo.FindByUID(uid)
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

// CreateProfileByUID - UID로 프로필 생성 또는 업데이트
func (s *userService) CreateProfileByUID(req dto.CreateProfileRequest) error {
	// UID string을 uuid.UUID로 파싱
	uid, err := uuid.Parse(req.UID)
	if err != nil {
		return errors.BadRequest("Invalid UID format")
	}

	// 1. User 찾기 또는 생성
	user, err := s.userRepo.FindByUID(uid)

	if err == gorm.ErrRecordNotFound {
		// User 없으면 생성
		newUser := &model.User{
			UID:          uid,
			RefreshToken: req.RefreshToken,
			Email:        req.UID + "@temp.com",
			Password:     "temp_password",
			UserRole:     "user",
			Nickname:     req.Nickname,
			Age:          req.Age,
			Gender:       req.Gender,
			Region:       req.Region,
			Bio:          req.Bio,
			Avatar:       req.Avatar,
		}

		if err := s.userRepo.Create(newUser); err != nil {
			return errors.WrapDatabase(err, "Failed to create user")
		}
		user = newUser
	} else if err != nil {
		return errors.WrapDatabase(err, "Failed to find user")
	} else {
		// User 있으면 업데이트
		user.RefreshToken = req.RefreshToken
		user.Nickname = req.Nickname
		user.Age = req.Age
		user.Gender = req.Gender
		user.Region = req.Region
		user.Bio = req.Bio
		user.Avatar = req.Avatar
		if err := s.userRepo.Update(user); err != nil {
			return errors.WrapDatabase(err, "Failed to update user")
		}
	}

	// 2. ProfileImage 처리
	if req.ProfileImage != "" {
		var file model.File
		if err := s.userRepo.DB().Where("file_url = ?", req.ProfileImage).First(&file).Error; err == nil {
			var existingProfileImage model.ProfileImage
			notFound := s.userRepo.DB().Where("uid = ? AND is_primary = ?", uid, true).First(&existingProfileImage).Error == gorm.ErrRecordNotFound

			if notFound {
				s.AddProfileImage(uid, file.ID, true)
			} else if existingProfileImage.FileID != file.ID {
				s.AddProfileImage(uid, file.ID, true)
			}
		}
	}

	return nil
}

// UpdateProfile - 프로필 수정
func (s *userService) UpdateProfile(uid uuid.UUID, req dto.UpdateProfileRequest) error {
	// User 찾기
	user, err := s.userRepo.FindByUID(uid)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NotFound("User not found")
		}
		return errors.WrapDatabase(err, "Failed to find user")
	}

	// 업데이트
	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}
	if req.Age > 0 {
		user.Age = req.Age
	}
	if req.Gender != "" {
		user.Gender = req.Gender
	}
	if req.Region != "" {
		user.Region = req.Region
	}
	user.Bio = req.Bio
	user.Avatar = req.Avatar

	if err := s.userRepo.Update(user); err != nil {
		return errors.WrapDatabase(err, "Failed to update user")
	}

	// ProfileImage 처리
	if req.ProfileImage != "" {
		var file model.File
		if err := s.userRepo.DB().Where("file_url = ?", req.ProfileImage).First(&file).Error; err == nil {
			var existingProfileImage model.ProfileImage
			notFound := s.userRepo.DB().Where("uid = ? AND is_primary = ?", uid, true).First(&existingProfileImage).Error == gorm.ErrRecordNotFound

			if notFound {
				s.AddProfileImage(uid, file.ID, true)
			} else if existingProfileImage.FileID != file.ID {
				s.AddProfileImage(uid, file.ID, true)
			}
		}
	}

	return nil
}

// GetUsersWithFilter - 필터링된 사용자 목록 조회
func (s *userService) GetUsersWithFilter(filter dto.UserListFilter) (*dto.UserListResponse, error) {
	users, total, err := s.userRepo.FindUsersWithFilter(filter)
	if err != nil {
		return nil, errors.WrapDatabase(err, "Failed to load users")
	}

	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &dto.UserListResponse{
		Users:      users,
		Total:      total,
		Page:       filter.Page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// AddProfileImage - 프로필 이미지 추가 (1:N 관계)
func (s *userService) AddProfileImage(uid uuid.UUID, fileID uint, isPrimary bool) error {
	// 1. User 존재 확인
	_, err := s.userRepo.FindByUID(uid)
	if err != nil {
		return errors.NotFound("User not found")
	}

	// 2. isPrimary가 true면 기존 primary 이미지들을 false로 변경
	if isPrimary {
		s.userRepo.DB().Model(&model.ProfileImage{}).
			Where("uid = ? AND is_primary = ?", uid, true).
			Update("is_primary", false)
	}

	// 3. 현재 프로필 이미지 개수 확인 (순서 결정용)
	var count int64
	s.userRepo.DB().Model(&model.ProfileImage{}).
		Where("uid = ?", uid).
		Count(&count)

	// 4. 새로운 프로필 이미지 추가
	profileImage := &model.ProfileImage{
		UID:       uid,
		FileID:    fileID,
		Order:     int(count) + 1,
		IsPrimary: isPrimary,
	}

	if err := s.userRepo.DB().Create(profileImage).Error; err != nil {
		return errors.WrapDatabase(err, "Failed to add profile image")
	}

	return nil
}
