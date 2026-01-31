package service

type FileService interface {
}

type fileService struct {
}

func NewFileService() FileService {
	return &fileService{}
}
