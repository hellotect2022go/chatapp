package middleware

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func RateLimitMiddleware(rdb *redis.Client, limit int, duration time.Duration) gin.HandlerFunc {
	//ctx := context.Background()
	return func(c *gin.Context) {
		// IP 주소 기반 (또는 사용자 ID)
		ctx := c.Request.Context()
		ip := c.ClientIP()
		key := fmt.Sprintf("rate_limit:%s", ip)

		// 현재 카운트 조회
		val, err := rdb.Get(ctx, key).Result()

		if err == redis.Nil {
			// 첫 요청
			rdb.Set(ctx, key, 1, duration)
			c.Next()
			return
		}

		if err != nil {
			c.JSON(500, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}

		count, _ := strconv.Atoi(val)

		if count >= limit {
			// 제한 초과
			c.JSON(429, gin.H{
				"error":   "Too many requests",
				"message": fmt.Sprintf("Rate limit exceeded. Try again in %v", duration),
			})
			c.Abort()
			return
		}

		// 카운트 증가
		rdb.Incr(ctx, key)
		c.Next()
	}
}

// 로그인 Rate Limit (1분에 5회)
func LoginRateLimit(rdb *redis.Client) gin.HandlerFunc {
	return RateLimitMiddleware(rdb, 5, 1*time.Minute)
}

// 회원가입 Rate Limit (1분에 3회)
func SignupRateLimit(rdb *redis.Client) gin.HandlerFunc {
	return RateLimitMiddleware(rdb, 3, 1*time.Minute)
}

// API Rate Limit (1분에 100회)
func APIRateLimit(rdb *redis.Client) gin.HandlerFunc {
	return RateLimitMiddleware(rdb, 100, 1*time.Minute)
}
