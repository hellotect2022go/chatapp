package repository

import (
	"github.com/google/uuid"
	"github.com/hellotect2022go/chatapp/internal/shared/model"
	"gorm.io/gorm"
)

type RoomRepository interface {
	Create(room *model.Room) error
	FindByID(roomID uuid.UUID) (*model.Room, error)
	FindByUserUID(userUID uuid.UUID) ([]*model.Room, error)
	FindAll() ([]*model.Room, error)
	FindDirectRoomByUsers(userUID1 uuid.UUID, userUID2 uuid.UUID) (*model.Room, error)
	AddMember(roomID uuid.UUID, userUID uuid.UUID) error
	RemoveMember(roomID uuid.UUID, userUID uuid.UUID) error
	GetMembers(roomID uuid.UUID) ([]*model.User, error)
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

func (r *roomRepository) FindByID(roomID uuid.UUID) (*model.Room, error) {
	var room model.Room
	// 채팅방 멤버 정보도 함께 조회
	err := r.db.Preload("Members").Where("room_id = ?", roomID).First(&room).Error

	if err != nil {
		return nil, err
	}

	return &room, nil
}

func (r *roomRepository) FindByUserUID(userUID uuid.UUID) ([]*model.Room, error) {
	var rooms []*model.Room

	// 내가 참여한 채팅방만 조회
	err := r.db.
		Joins("JOIN room_members ON rooms.room_id = room_members.room_id").
		Where("room_members.user_uid = ?", userUID).
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
		Order("created_at ASC").
		Find(&rooms).Error

	if err != nil {
		return nil, err
	}

	return rooms, nil
}

func (r *roomRepository) AddMember(roomID uuid.UUID, userUID uuid.UUID) error {
	member := &model.RoomMember{RoomID: roomID, UserUID: userUID}

	return r.db.Create(member).Error
}

func (r *roomRepository) RemoveMember(roomID uuid.UUID, userUID uuid.UUID) error {
	return r.db.
		Where("room_id = ? AND user_uid = ?", roomID, userUID).
		Delete(&model.RoomMember{}).Error
}

func (r *roomRepository) GetMembers(roomID uuid.UUID) ([]*model.User, error) {
	var users []*model.User
	err := r.db.Joins("JOIN room_members ON users.uid = room_members.user_uid").
		Where("room_members.room_id = ?", roomID).
		Find(&users).Error

	if err != nil {
		return nil, err
	}
	return users, nil
}

// FindDirectRoomByUsers - 두 사용자 간의 Direct 채팅방 찾기
func (r *roomRepository) FindDirectRoomByUsers(userUID1 uuid.UUID, userUID2 uuid.UUID) (*model.Room, error) {
	var room model.Room
	
	// Direct 타입이면서 두 사용자가 모두 멤버인 방 찾기
	err := r.db.
		Where("type = ?", model.RoomTypeDirect).
		Joins("JOIN room_members rm1 ON rooms.room_id = rm1.room_id AND rm1.user_uid = ?", userUID1).
		Joins("JOIN room_members rm2 ON rooms.room_id = rm2.room_id AND rm2.user_uid = ?", userUID2).
		// 정확히 2명만 있는 방 확인
		Where("rooms.room_id IN (SELECT room_id FROM room_members GROUP BY room_id HAVING COUNT(*) = 2)").
		First(&room).Error

	if err != nil {
		return nil, err
	}

	return &room, nil
}
