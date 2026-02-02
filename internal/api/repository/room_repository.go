package repository

import (
	"github.com/hellotect2022go/chatapp/internal/shared/model"
	"gorm.io/gorm"
)

type RoomRepository interface {
	Create(room *model.Room) error
	FindByID(id uint) (*model.Room, error)
	FindByUserID(userID uint) ([]*model.Room, error)
	FindAll() ([]*model.Room, error) // ⭐ 추가: Redis 복원용
	FindDirectRoomByUsers(userID1 uint, userID2 uint) (*model.Room, error) // ⭐ 추가: Direct 방 중복 확인
	AddMember(roomID uint, userID uint) error
	RemoveMember(roomID uint, userID uint) error
	GetMembers(roomID uint) ([]*model.User, error)
}

type roomRepository struct {
	db *gorm.DB
}

func NewRoomRepository(db *gorm.DB) RoomRepository {
	return &roomRepository{db: db}
}

func (r *roomRepository) Create(room *model.Room) error {
	return r.db.Create(room).Error
}

func (r *roomRepository) FindByID(id uint) (*model.Room, error) {
	var room model.Room
	// 채팅방 멤버 정보도 함께 조회
	err := r.db.Preload("Members").First(&room, id).Error

	if err != nil {
		return nil, err
	}

	return &room, nil
}

func (r *roomRepository) FindByUserID(userID uint) ([]*model.Room, error) {
	var rooms []*model.Room

	// 내가 참여한 채팅방만 조회
	err := r.db.
		Joins("JOIN room_members ON rooms.id = room_members.room_id").
		Where("room_members.user_id = ?", userID).
		Order("rooms.updated_at DESC").
		Find(&rooms).Error

	if err != nil {
		return nil, err
	}

	return rooms, nil
}

func (r *roomRepository) FindAll() ([]*model.Room, error) {
	var rooms []*model.Room

	err := r.db.
		Order("id ASC").
		Find(&rooms).Error

	if err != nil {
		return nil, err
	}

	return rooms, nil
}

func (r *roomRepository) AddMember(roomID uint, userID uint) error {
	member := &model.RoomMember{RoomID: roomID, UserID: userID}

	return r.db.Create(member).Error
}

func (r *roomRepository) RemoveMember(roomID uint, userID uint) error {
	return r.db.
		Where("room_id = ? AND user_id = ?", roomID, userID).
		Delete(&model.RoomMember{}).Error
}

func (r *roomRepository) GetMembers(roomID uint) ([]*model.User, error) {
	var users []*model.User
	err := r.db.Joins("JOIN room_members ON users.id = room_members.user_id").
		Where("room_members.room_id = ?", roomID).
		Find(&users).Error

	if err != nil {
		return nil, err
	}
	return users, nil
}

// ⭐ FindDirectRoomByUsers - 두 사용자 간의 Direct 채팅방 찾기
func (r *roomRepository) FindDirectRoomByUsers(userID1 uint, userID2 uint) (*model.Room, error) {
	var room model.Room
	
	// Direct 타입이면서 두 사용자가 모두 멤버인 방 찾기
	err := r.db.
		Where("type = ?", model.RoomTypeDirect).
		Joins("JOIN room_members rm1 ON rooms.id = rm1.room_id AND rm1.user_id = ?", userID1).
		Joins("JOIN room_members rm2 ON rooms.id = rm2.room_id AND rm2.user_id = ?", userID2).
		// 정확히 2명만 있는 방 확인
		Where("rooms.id IN (SELECT room_id FROM room_members GROUP BY room_id HAVING COUNT(*) = 2)").
		First(&room).Error

	if err != nil {
		return nil, err
	}

	return &room, nil
}
