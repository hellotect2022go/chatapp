package errors

import (
	"fmt"
	"net/http"
)

type ErrorCode string

const (
	// 인증 관련 에러
	ErrUnauthorized    ErrorCode = "UNAUTHORIZED"
	ErrInvalidToken    ErrorCode = "INVALID_TOKEN"
	ErrTokenExpired    ErrorCode = "TOKEN_EXPIRED"
	ErrInvalidPassword ErrorCode = "INVALID_PASSWORD"

	// 리소스 관련 에러
	ErrNotFound      ErrorCode = "NOT_FOUND"
	ErrAlreadyExists ErrorCode = "ALREADY_EXISTS"
	ErrForbidden     ErrorCode = "FORBIDDEN"

	// 요청 관련 에러
	ErrBadRequest ErrorCode = "BAD_REQUEST"
	ErrValidation ErrorCode = "VALIDATION_ERROR"

	// 서버 관련 에러
	ErrInternal ErrorCode = "INTERNAL_ERROR"
	ErrDatabase ErrorCode = "DATABASE_ERROR"
	ErrRedis    ErrorCode = "REDIS_ERROR"

	// 파일 관련 에러
	ErrFileTooBig       ErrorCode = "FILE_TOO_BIG"
	ErrInvalidFileType  ErrorCode = "INVALID_FILE_TYPE"
	ErrFileUploadFailed ErrorCode = "FILE_UPLOAD_FAILED"
)

// AppError - 애플리케이션 에러 구조체
type AppError struct {
	Code       ErrorCode `json:"code"`
	Message    string    `json:"message"`
	HTTPStatus int       `json:"-"`
	Err        error     `json:"-"` // 원본 에러 (로깅용)
}

// Error - error 인터페이스 구현
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (original: %v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap - 원본 에러 반환
func (e *AppError) Unwrap() error {
	return e.Err
}

// ============================================================
// 에러 생성 헬퍼 함수
// ============================================================

// New - 새로운 AppError 생성
func New(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: getHTTPStatus(code),
	}
}

// Wrap - 기존 에러를 AppError로 래핑
func Wrap(err error, code ErrorCode, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: getHTTPStatus(code),
		Err:        err,
	}
}

// ============================================================
// 자주 사용하는 에러들
// ============================================================

func Unauthorized(message string) *AppError {
	return New(ErrUnauthorized, message)
}

func NotFound(resource string) *AppError {
	return New(ErrNotFound, fmt.Sprintf("%s not found", resource))
}

func AlreadyExists(resource string) *AppError {
	return New(ErrAlreadyExists, fmt.Sprintf("%s already exists", resource))
}

func BadRequest(message string) *AppError {
	return New(ErrBadRequest, message)
}

func Forbidden(message string) *AppError {
	return New(ErrForbidden, message)
}

func Internal(message string) *AppError {
	return New(ErrInternal, message)
}

func WrapInternal(err error, message string) *AppError {
	return Wrap(err, ErrInternal, message)
}

func WrapDatabase(err error, message string) *AppError {
	return Wrap(err, ErrDatabase, message)
}

func ValidationError(message string) *AppError {
	return New(ErrValidation, message)
}

// ============================================================
// HTTP 상태 코드 매핑
// ============================================================

func getHTTPStatus(code ErrorCode) int {
	switch code {
	case ErrUnauthorized, ErrInvalidToken, ErrTokenExpired, ErrInvalidPassword:
		return http.StatusUnauthorized
	case ErrNotFound:
		return http.StatusNotFound
	case ErrAlreadyExists:
		return http.StatusConflict
	case ErrForbidden:
		return http.StatusForbidden
	case ErrBadRequest, ErrValidation:
		return http.StatusBadRequest
	case ErrFileTooBig, ErrInvalidFileType, ErrFileUploadFailed:
		return http.StatusBadRequest
	case ErrDatabase, ErrRedis, ErrInternal:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// ============================================================
// 에러 변환 유틸리티
// ============================================================

// IsAppError - AppError 타입 체크
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// GetAppError - AppError 추출
func GetAppError(err error) *AppError {
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}
	return nil
}
