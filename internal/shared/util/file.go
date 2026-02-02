package util

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ValidateImage - 이미지 파일 검증
func ValidateImage(file multipart.File, size int64) (string, error) {
	// 1. 크기 검증 (10MB 제한)
	if size > 10*1024*1024 {
		return "", errors.New("file size exceeds 10MB")
	}

	// 2. MIME 타입 검증
	buffer := make([]byte, 512)
	_, err := file.Read(buffer)
	if err != nil {
		return "", err
	}
	file.Seek(0, 0) // 파일 포인터 초기화

	contentType := http.DetectContentType(buffer)

	switch contentType {
	case "image/jpeg", "image/jpg":
		return "image/jpeg", nil
	case "image/png":
		return "image/png", nil
	case "image/gif":
		return "image/gif", nil
	case "image/webp":
		return "image/webp", nil
	default:
		return "", fmt.Errorf("invalid image type: %s", contentType)
	}
}

// GenerateSafeFilename - 안전한 파일명 생성
func GenerateSafeFilename(originalName string) string {
	ext := strings.ToLower(filepath.Ext(originalName))
	uniqueID := uuid.New().String()
	return fmt.Sprintf("%s%s", uniqueID, ext)
}

// SaveFile - 로컬에 파일 저장
func SaveFile(file multipart.File, filename string) (string, error) {
	// 날짜별 디렉토리 생성 (step11 디렉토리 기준)
	dir := filepath.Join("uploads", time.Now().Format("2006/01/02"))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	// 파일 경로 (로컬 저장용 - OS별 구분자 사용)
	filePath := filepath.Join(dir, filename)

	// 파일 생성
	dst, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer dst.Close() // ⭐ 함수 종료 시 닫기

	// 파일 내용 복사
	if _, err := io.Copy(dst, file); err != nil {
		return "", err
	}

	// ⭐ URL용 경로 반환 (항상 슬래시 사용)
	urlPath := strings.ReplaceAll(filePath, "\\", "/")
	return urlPath, nil
}
