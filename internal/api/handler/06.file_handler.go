package handler

import (
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

// POST /api/v1/upload/image - 프로필 이미지 업로드
func (h *FileHandler) UploadProfileImage(c *gin.Context) {
	// 1. 파일 가져오기
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error": "No file uploaded"})
		return
	}
	defer file.Close()

	// 2. UID 가져오기 (폼데이터에서)
	uidStr := c.PostForm("userId")
	if uidStr == "" {
		c.JSON(400, gin.H{"error": "UID required"})
		return
	}

	// 3. UID string을 uuid.UUID로 파싱
	uid, err := uuid.Parse(uidStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid UID format"})
		return
	}

	log.Printf("📸 프로필 이미지 업로드: uid=%s, filename=%s, size=%d",
		uid.String(), header.Filename, header.Size)

	// 4. 이미지 업로드 (uuid.UUID 전달)
	uploadedFile, err := h.fileService.UploadImage(file, header, uid)
	if err != nil {
		log.Printf("❌ 이미지 업로드 실패: %v", err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	log.Printf("✅ 이미지 업로드 성공: %s", uploadedFile.FileURL)

	// 5. 성공 응답
	c.JSON(200, gin.H{
		"success":  true,
		"url":      uploadedFile.FileURL,
		"imageUrl": uploadedFile.FileURL,
		"file":     uploadedFile,
	})
}
