package repository

import (
	"github.com/hellotect2022go/chatapp/internal/shared/model"
	"gorm.io/gorm"
)

type MessageRepository interface {
	Create(message *model.Message) error
	FindByRoomID(roomID uint, limit, offset int) ([]*model.Message, error)
	CountByRoomID(roomID uint) (int64, error)
}

type messageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) Create(message *model.Message) error {
	panic("not implemented") // TODO: Implement
}

func (r *messageRepository) FindByRoomID(roomID uint, limit int, offset int) ([]*model.Message, error) {
	panic("not implemented") // TODO: Implement
}

func (r *messageRepository) CountByRoomID(roomID uint) (int64, error) {
	panic("not implemented") // TODO: Implement
}
