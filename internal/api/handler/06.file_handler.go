package handler

import (
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hellotect2022go/chatapp/internal/api/service"
)

type FileHandler struct {
	fileService service.FileService
}

func NewFileHandler(fileService service.FileService) *FileHandler {
	return &FileHandler{
		fileService: fileService,
	}
}

// GET /api/v1/files/:id - 파일 다운로드 (권한 체크 포함)
func (h *FileHandler) DownloadFile(c *gin.Context) {
	fileID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid file ID"})
		return
	}

	// 1. 파일 정보 조회
	file, err := h.fileService.GetFileByID(uint(fileID))
	if err != nil {
		c.JSON(404, gin.H{"error": "File not found"})
		return
	}

	// 2. 권한 체크 (선택적)
	// TODO: 메시지 접근 권한 확인
	// - 해당 메시지가 속한 채팅방의 멤버인지 확인
	// - 현재는 모든 인증된 사용자에게 허용

	// 3. 파일 제공
	log.Printf("Serving file: %s (Original: %s)", file.FilePath, file.OriginalName)

	// Content-Type 설정
	c.Header("Content-Type", file.MimeType)
	// 파일명 설정 (다운로드 시 원본 파일명 사용)
	c.Header("Content-Disposition", "inline; filename=\""+file.OriginalName+"\"")

	// 파일 전송
	c.File(file.FilePath)
}
