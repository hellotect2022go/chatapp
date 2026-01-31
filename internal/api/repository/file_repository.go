package repository

import (
	"github.com/hellotect2022go/chatapp/internal/shared/model"
	"gorm.io/gorm"
)

type FileRepository interface {
	Create(file *model.File) error
	FindByID(id uint) (*model.File, error)
	FindByMessageID(messageID uint) ([]*model.File, error)
	Delete(id uint) error
}

type fileRepository struct {
	db *gorm.DB
}

func NewFileRepository(db *gorm.DB) FileRepository {
	return &fileRepository{db: db}
}

func (r *fileRepository) Create(file *model.File) error {
	panic("not implemented") // TODO: Implement
}

func (r *fileRepository) FindByID(id uint) (*model.File, error) {
	panic("not implemented") // TODO: Implement
}

func (r *fileRepository) FindByMessageID(messageID uint) ([]*model.File, error) {
	panic("not implemented") // TODO: Implement
}

func (r *fileRepository) Delete(id uint) error {
	panic("not implemented") // TODO: Implement
}
