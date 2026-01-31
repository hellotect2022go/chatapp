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
	panic("not implemented") // TODO: Implement
}

func (r *roomRepository) FindByID(id uint) (*model.Room, error) {
	panic("not implemented") // TODO: Implement
}

func (r *roomRepository) FindByUserID(userID uint) ([]*model.Room, error) {
	panic("not implemented") // TODO: Implement
}

func (r *roomRepository) FindAll() ([]*model.Room, error) {
	panic("not implemented") // TODO: Implement
}

func (r *roomRepository) AddMember(roomID uint, userID uint) error {
	panic("not implemented") // TODO: Implement
}

func (r *roomRepository) RemoveMember(roomID uint, userID uint) error {
	panic("not implemented") // TODO: Implement
}

func (r *roomRepository) GetMembers(roomID uint) ([]*model.User, error) {
	panic("not implemented") // TODO: Implement
}
