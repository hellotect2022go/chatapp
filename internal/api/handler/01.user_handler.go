package handler

import (
	"fmt"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hellotect2022go/chatapp/internal/api/dto"
	"github.com/hellotect2022go/chatapp/internal/api/service"
	"github.com/hellotect2022go/chatapp/internal/shared/logger"
	"go.uber.org/zap"
)

type UserHandler struct {
	userService service.UserService
	fileService service.FileService
}

func NewUserHandler(userService service.UserService, fileService service.FileService) *UserHandler {
	return &UserHandler{
		userService: userService,
		fileService: fileService,
	}
}

func (h *UserHandler) GetMe(c *gin.Context) {
	uidVal, exists := c.Get("uid")
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}
	uid, err := uuid.Parse(uidVal.(string))
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid uid format"})
		return
	}

	// 2. Service 호출
	user, err := h.userService.GetByUID(uid)
	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, user)
}

func (h *UserHandler) UpdateMe(c *gin.Context) {
	uidVal, exists := c.Get("uid")
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}
	uid, err := uuid.Parse(uidVal.(string))
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid uid format"})
		return
	}

	var req dto.UpdateMeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 2. Service 호출
	if err := h.userService.Update(uid, req.Nickname, req.Age); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "User updated successfully"})
}

func (h *UserHandler) DeleteMe(c *gin.Context) {
	uidVal, exists := c.Get("uid")
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}
	uid, err := uuid.Parse(uidVal.(string))
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid uid format"})
		return
	}
	// 2. Service 호출
	if err := h.userService.Delete(uid); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "User deleted successfully"})
}

func (h *UserHandler) ChangePassword(c *gin.Context) {
	uidVal, exists := c.Get("uid")
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}
	uid, err := uuid.Parse(uidVal.(string))
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid uid format"})
		return
	}
	var req dto.ChangePasswordRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	// 2. Service 호출
	if err := h.userService.ChangePassword(uid, req.Password); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Password changed successfully"})
}

// Admin Handlers
func (h *UserHandler) GetAll(c *gin.Context) {
	users, err := h.userService.GetAll()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, users)

}

func (h *UserHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")

	uid, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid ID"})
		return
	}

	if err := h.userService.Delete(uid); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "User deleted successfully"})
}

// GET /api/v1/users/recent - 최근 로그인한 사용자 목록
func (h *UserHandler) GetRecentUsers(c *gin.Context) {
	// limit 쿼리 파라미터 (기본값: 20)
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}

	users, err := h.userService.GetRecentUsers(limit)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"users": users,
		"total": len(users),
	})
}

// POST /api/v1/users/profile - 프로필 생성
func (h *UserHandler) CreateProfile(c *gin.Context) {
	var req dto.CreateProfileRequest

	// JSON 바인딩 (이미지 URL은 이미 업로드된 상태로 전달됨)
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("JSON 파싱 에러", zap.Error(err))
		c.JSON(400, gin.H{"error": "데이터 형식이 올바르지 않습니다."})
		return
	}

	log.Printf("📝 프로필 생성 요청: %+v", req)

	// UID가 없으면 에러
	if req.UID == "" {
		c.JSON(400, gin.H{"error": "UID is required"})
		return
	}

	// 4. 유효성 검사
	if req.Nickname == "" {
		c.JSON(400, gin.H{"error": "Nickname is required"})
		return
	}
	if req.Age < 19 {
		c.JSON(400, gin.H{"error": "Age must be at least 19"})
		return
	}
	if req.Gender == "" {
		c.JSON(400, gin.H{"error": "Gender is required"})
		return
	}
	if req.Region == "" {
		c.JSON(400, gin.H{"error": "Region is required"})
		return
	}

	// 5. UID로 사용자 생성 (인증 없이 바로 프로필 생성 가능)
	if err := h.userService.CreateProfileByUID(req); err != nil {
		log.Printf("❌ 프로필 생성 실패: %v", err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// 6. 생성된 사용자 정보 반환
	uid, _ := uuid.Parse(req.UID)
	user, _ := h.userService.GetByUID(uid)

	log.Printf("✅ 프로필 생성 완료: uid=%s", user.UID.String())

	c.JSON(200, gin.H{
		"success": true,
		"user":    user,
		"message": "Profile created successfully",
	})
}

// PUT /api/v1/users/profile - 프로필 수정
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	var req dto.UpdateProfileRequest

	// UID로 처리 (미들웨어에서 uid를 설정하거나, 요청에서 받음)
	uidVal, exists := c.Get("uid")

	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Println(err)
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if !exists {
		// 미들웨어에 uid가 없으면 요청 body에서 가져옴
		uidVal = req.UID
	}

	fmt.Println("uid : ", uidVal)

	uidStr, ok := uidVal.(string)
	if !ok {
		c.JSON(400, gin.H{"error": "Invalid uid type"})
		return
	}
	uid, err := uuid.Parse(uidStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid uid format"})
		return
	}

	if err := h.userService.UpdateProfile(uid, req); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// 업데이트된 사용자 정보 반환
	user, _ := h.userService.GetByUID(uid)

	c.JSON(200, gin.H{
		"success": true,
		"user":    user,
		"message": "Profile updated successfully",
	})
}

// GET /api/v1/users - 사용자 목록 조회 (필터링)
func (h *UserHandler) GetUsersList(c *gin.Context) {
	var filter dto.UserListFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	result, err := h.userService.GetUsersWithFilter(filter)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, result)
}
