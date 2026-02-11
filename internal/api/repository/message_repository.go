package repository

import (
	"github.com/google/uuid"
	"github.com/hellotect2022go/chatapp/internal/shared/model"
	"gorm.io/gorm"
)

type MessageRepository interface {
	Create(message *model.Message) error
	FindByRoomID(roomID uuid.UUID, limit, offset int) ([]*model.Message, error)
	CountByRoomID(roomID uuid.UUID) (int64, error)
}

type messageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) Create(message *model.Message) error {
	return r.db.Create(message).Error
}

func (r *messageRepository) FindByRoomID(roomID uuid.UUID, limit int, offset int) ([]*model.Message, error) {
	var messages []*model.Message

	err := r.db.
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("uid, nickname")
		}).
		Preload("Files").
		Where("room_id = ?", roomID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&messages).Error

	return messages, err
}

func (r *messageRepository) CountByRoomID(roomID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&model.Message{}).
		Where("room_id = ?", roomID).
		Count(&count).Error
	return count, err
}
