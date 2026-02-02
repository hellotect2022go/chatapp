package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/hellotect2022go/chatapp/internal/shared/errors"
	"github.com/hellotect2022go/chatapp/internal/shared/logger"
	"go.uber.org/zap"
)

// ErrorHandler - 에러 핸들링 미들웨어
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 다음 핸들러 실행
		c.Next()

		// 에러가 있는지 확인
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err

			// AppError인 경우
			if appErr := errors.GetAppError(err); appErr != nil {
				// 서버 에러는 상세 로깅 (원본 에러 포함)
				if appErr.HTTPStatus >= 500 {
					logger.Error("Application error",
						zap.String("code", string(appErr.Code)),
						zap.String("message", appErr.Message),
						zap.Int("status", appErr.HTTPStatus),
						zap.String("path", c.Request.URL.Path),
						zap.String("method", c.Request.Method),
						zap.Error(appErr.Err), // 원본 에러
					)
				} else {
					// 클라이언트 에러는 간단히 로깅
					logger.Warn("Client error",
						zap.String("code", string(appErr.Code)),
						zap.String("message", appErr.Message),
						zap.Int("status", appErr.HTTPStatus),
						zap.String("path", c.Request.URL.Path),
					)
				}

				// 클라이언트에게 응답
				c.JSON(appErr.HTTPStatus, gin.H{
					"error": gin.H{
						"code":    appErr.Code,
						"message": appErr.Message,
					},
				})
				return
			}

			// 일반 에러인 경우 (Internal Server Error로 처리)
			logger.Error("Unhandled error",
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
				zap.Error(err),
			)

			c.JSON(500, gin.H{
				"error": gin.H{
					"code":    "INTERNAL_ERROR",
					"message": "Internal server error",
				},
			})
		}
	}
}

// Recovery - Panic 복구 미들웨어 (커스텀)
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("Panic recovered",
					zap.String("method", c.Request.Method),
					zap.String("path", c.Request.URL.Path),
					zap.Any("panic", err),
					zap.Stack("stack"),
				)

				c.JSON(500, gin.H{
					"error": gin.H{
						"code":    "INTERNAL_ERROR",
						"message": "Internal server error",
					},
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}
