package service

import (
	"fmt"
	"log"

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

type UserService interface {
	GetByUID(uid uuid.UUID) (*model.User, error)
	Update(uid uuid.UUID, nickname string, age int) error
	Delete(uid uuid.UUID) error
	ChangePassword(uid uuid.UUID, newPassword string) error
	GetAll() ([]*model.User, error)
	GetRecentUsers(limit int) ([]*model.User, error)
	CreateProfileByUID(req dto.CreateProfileRequest) (*dto.AuthResponse, error)
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
func (s *userService) CreateProfileByUID(req dto.CreateProfileRequest) (*dto.AuthResponse, error) {
	// UID string을 uuid.UUID로 파싱
	uid, err := uuid.Parse(req.UID)
	if err != nil {
		return nil, errors.BadRequest("Invalid UID format")
	}

	// 1. User 찾기 또는 생성
	user, err := s.userRepo.FindByUID(uid)

	// 3. ⭐ JWT 토큰 생성
	token, err := util.GenerateToken(user.UID, user.Nickname, user.UserRole)
	if err != nil {
		return nil, fmt.Errorf("Failed to generate token")
	}

	if err == gorm.ErrRecordNotFound {
		// User 없으면 생성
		newUser := &model.User{
			UID:          uid,
			RefreshToken: token.RefreshToken,
			Email:        req.UID + "@temp.com",
			Password:     "temp_password",
			Nickname:     req.Nickname,
			Age:          req.Age,
			Gender:       req.Gender,
			Region:       req.Region,
			Bio:          req.Bio,
			Avatar:       req.Avatar,
			IsCompleted:  true,
		}

		if err := s.userRepo.Create(newUser); err != nil {
			return nil, errors.WrapDatabase(err, "Failed to create user")
		}
		user = newUser
	} else if err != nil {
		return nil, errors.WrapDatabase(err, "Failed to find user")
	} else {
		// User 있으면 업데이트 (FirebaseAuth에서 생성된 유저 프로필 완료)
		user.Nickname = req.Nickname
		user.RefreshToken = token.RefreshToken
		user.Age = req.Age
		user.Gender = req.Gender
		user.Region = req.Region
		user.Bio = req.Bio
		user.Avatar = req.Avatar
		user.IsCompleted = true

		if err := s.userRepo.Update(user); err != nil {
			return nil, errors.WrapDatabase(err, "Failed to update user")
		}
	}

	// 2. ProfileImage 처리
	if len(req.ProfileImages) > 0 {
		// 혹시 해당 uid 랑 연계 되어있는 ProfileImage 전체 삭제
		s.userRepo.DB().Where("uid = ?", uid).Delete(&model.ProfileImage{})

		// 각 url 에 대해 반복
		for i, fileURL := range req.ProfileImages {
			// File 테이블에서 File 찾기
			var file model.File
			if err := s.userRepo.DB().Where("file_url = ?", fileURL).First(&file).Error; err != nil {
				logger.Error("", zap.Error(err))
				continue
			}

			profileImage := model.ProfileImage{
				UID:       uid,
				FileID:    file.ID,
				Order:     i,
				IsPrimary: i == 0,
			}

			if err := s.userRepo.DB().Create(&profileImage).Error; err != nil {
				// 에러 처리
				return nil, errors.WrapDatabase(err, "Failed to create profile image")
			}

		}
	}

	return &dto.AuthResponse{
		User:            *user,
		AccessToken:     token.AccessToken,
		RefreshToken:    token.RefreshToken,
		ProfileComplete: true, // 클라이언트는 메인 화면으로 이동
	}, nil
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

	// 2. User 기본 정보 업데이트
	user.Nickname = req.Nickname
	user.Age = req.Age
	user.Gender = req.Gender
	user.Region = req.Region
	user.Bio = req.Bio
	user.Avatar = req.Avatar

	if err := s.userRepo.Update(user); err != nil {
		return errors.WrapDatabase(err, "Failed to update user")
	}

	// 1. 삭제 요청이 들어온 이미지만 타켓팅 해서 삭제
	if len(req.DeletedImages) > 0 {
		// File URL 을 통해 관련 ProfileImage 삭제
		// 1. 삭제할 FileID들을 서브쿼리로 정의
		subQuery := s.userRepo.DB().Model(&model.File{}).
			Select("id").
			Where("file_url IN ?", req.DeletedImages)

		// 2. ProfileImage에서 해당 FileID를 가진 레코드 삭제
		err := s.userRepo.DB().
			Where("uid = ? AND file_id IN (?)", uid, subQuery).
			Delete(&model.ProfileImage{}).Error

		if err != nil {
			return errors.WrapDatabase(err, "Failed to delete specific profile images")
		}
	}

	// 2. 현재 요청된 순서대로 ProfileImage 상태 업데이트 또는 생성
	for i, imageUrl := range req.ProfileImages {
		var file model.File
		if err := s.userRepo.DB().Where("file_url = ?", imageUrl).First(&file).Error; err != nil {
			continue
		}

		// Upsert (있으면 업데이트, 없으면 생성)
		// GORM의 Save나 Clauses(OnConflict)를 사용하거나,
		// 단순하게는 기존에 있는지 체크 후 처리
		profileImage := model.ProfileImage{
			UID:    uid,
			FileID: file.ID,
		}

		// Assign으로 변경될 값 설정
		err := s.userRepo.DB().Where(profileImage).Assign(model.ProfileImage{
			Order:     i,
			IsPrimary: i == 0,
		}).FirstOrCreate(&profileImage).Error

		if err != nil {
			return errors.WrapDatabase(err, "Failed to sync profile image")
		}
	}

	// ⭐ 1. 삭제할 이미지가 있으면 먼저 처리
	if len(req.DeletedImages) > 0 {
		log.Printf("🗑️ 삭제 요청된 이미지: %v", req.DeletedImages)
		if err := s.userRepo.DeleteProfileImagesByURLs(uid, req.DeletedImages); err != nil {
			return errors.WrapDatabase(err, "Failed to delete images")
		}
	}
	// // 3. ProfileImage 처리
	// if len(req.ProfileImages) > 0 {
	// 	// 기존 ProfileImage 모두 삭제
	// 	if err := s.userRepo.DB().Where("uid = ?", uid).Delete(&model.ProfileImage{}).Error; err != nil {
	// 		return errors.WrapDatabase(err, "Failed to delete old profile images")
	// 	}

	// 	// 새 ProfileImage 레코드 생성
	// 	for i, imageUrl := range req.ProfileImages {
	// 		var file model.File
	// 		if err := s.userRepo.DB().Where("file_url = ?", imageUrl).First(&file).Error; err != nil {
	// 			log.Printf("⚠️ File not found for URL: %s", imageUrl)
	// 			continue
	// 		}

	// 		profileImage := model.ProfileImage{
	// 			UID:       uid,
	// 			FileID:    file.ID,
	// 			Order:     i,
	// 			IsPrimary: i == 0,
	// 		}

	// 		if err := s.userRepo.DB().Create(&profileImage).Error; err != nil {
	// 			return errors.WrapDatabase(err, "Failed to create profile image")
	// 		}
	// 	}

	// 	log.Printf("✅ ProfileImage 업데이트 완료: %d개", len(req.ProfileImages))
	// }

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
