package middleware

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/hellotect2022go/chatapp/internal/shared/util"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(401, gin.H{"error": "Unauthorized"})
			c.Abort() // 중요 다음 핸들러 실행 중단
			return
		}

		parts := strings.Split(token, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(401, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := parts[1]
		// 토큰 유효성 검증
		claims, err := util.ValidateToken(tokenString)
		if err != nil {
			c.JSON(401, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		fmt.Println(claims)

		c.Set("uid", claims.UserUID)      // ⭐ JWT에서 받은 string UID를 그대로 저장
		c.Set("nickname", claims.Nickname)
		c.Set("user_role", claims.UserRole)
		c.Next()
	}
}

func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 앞선 AuthMiddleware에서 저장한 정보 꺼내기
		// (보통 AuthMiddleware가 먼저 실행되어야 합니다)
		userRole, exists := c.Get("user_role")

		if !exists {
			c.JSON(401, gin.H{"error": "인증 정보가 없습니다."})
			c.Abort() // 중요: 이후 핸들러 실행을 막음
			return
		}

		// 2. 권한 체크
		if userRole != "admin" {
			c.JSON(403, gin.H{"error": "관리자만 접근 가능합니다."})
			c.Abort() // 중요: 여기서 중단
			return
		}

		// 3. 관리자라면 다음 단계로 진행
		c.Next()
	}
}
