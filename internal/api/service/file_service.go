package service

import (
	"fmt"
	"mime/multipart"

	"github.com/hellotect2022go/chatapp/internal/api/repository"
	"github.com/hellotect2022go/chatapp/internal/shared/errors"
	"github.com/hellotect2022go/chatapp/internal/shared/model"
	"github.com/hellotect2022go/chatapp/internal/shared/util"
)

type FileService interface {
	UploadImage(file multipart.File, header *multipart.FileHeader, messageID uint) (*model.File, error)
	GetFileURL(filepath string) string
	GetFileByID(fileID uint) (*model.File, error)
}

type fileService struct {
	fileRepo repository.FileRepository
	baseURL  string
}

func NewFileService(fileRepo repository.FileRepository, baseURL string) FileService {
	return &fileService{
		fileRepo: fileRepo,
		baseURL:  baseURL,
	}
}

func (s *fileService) UploadImage(file multipart.File, header *multipart.FileHeader, messageID uint) (*model.File, error) {
	// 1. 이미지 검증
	mimeType, err := util.ValidateImage(file, header.Size)
	if err != nil {
		// util.ValidateImage에서 발생하는 에러는 이미 명확하므로 그대로 전달
		return nil, errors.BadRequest(err.Error())
	}

	// 2. 안전한 파일명 생성 (UUID)
	safeFilename := util.GenerateSafeFilename(header.Filename)

	// 3. 파일 저장
	filepath, err := util.SaveFile(file, safeFilename)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrFileUploadFailed, "Failed to save file")
	}

	// 4. URL 생성 (정적 파일 서빙 경로)
	fileURL := s.GetFileURL(filepath)

	// 5. DB 저장
	fileModel := &model.File{
		MessageID:    messageID,
		OriginalName: header.Filename,
		StoredName:   safeFilename,
		FilePath:     filepath,
		FileURL:      fileURL,
		MimeType:     mimeType,
		FileSize:     header.Size,
	}

	if err := s.fileRepo.Create(fileModel); err != nil {
		return nil, errors.WrapDatabase(err, "Failed to save file metadata")
	}

	return fileModel, nil
}

func (s *fileService) GetFileURL(filepath string) string {
	// ⭐ 정적 파일 서빙 URL (브라우저 <img> 태그에서 직접 사용)
	return fmt.Sprintf("%s/%s", s.baseURL, filepath)
}

// ⭐ API를 통한 파일 접근 (권한 체크 필요 시 사용)
func (s *fileService) GetFileURLByID(fileID uint) string {
	return fmt.Sprintf("%s/api/v1/files/%d", s.baseURL, fileID)
}

func (s *fileService) GetFileByID(fileID uint) (*model.File, error) {
	return s.fileRepo.FindByID(uint(fileID))
}
