package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hellotect2022go/chatapp/internal/shared/logger"
	"go.uber.org/zap"
)

// Logger - zap 기반 로깅 미들웨어
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// 다음 핸들러 실행
		c.Next()

		// 응답 후 로깅
		latency := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		// 로그 레벨 결정
		switch {
		case statusCode >= 500:
			logger.Error("Server error",
				zap.Int("status", statusCode),
				zap.String("method", method),
				zap.String("path", path),
				zap.String("query", query),
				zap.String("ip", clientIP),
				zap.Duration("latency", latency),
				zap.String("error", errorMessage),
			)
		case statusCode >= 400:
			logger.Warn("Client error",
				zap.Int("status", statusCode),
				zap.String("method", method),
				zap.String("path", path),
				zap.String("query", query),
				zap.String("ip", clientIP),
				zap.Duration("latency", latency),
			)
		default:
			logger.Info("Request",
				zap.Int("status", statusCode),
				zap.String("method", method),
				zap.String("path", path),
				zap.String("ip", clientIP),
				zap.Duration("latency", latency),
			)
		}
	}
}
