package handler

import (
	"magang-be/services/magang/internal/service"
	"magang-be/services/magang/pkg/auth"
	"magang-be/services/magang/pkg/response"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	userSvc   *service.UserService
	jwtSecret string
}

func NewAuthHandler(userSvc *service.UserService, jwtSecret string) *AuthHandler {
	return &AuthHandler{userSvc: userSvc, jwtSecret: jwtSecret}
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.Username == "" || req.Password == "" {
		response.BadRequest(c, "username and password are required")
		return
	}

	user, err := h.userSvc.GetByUsername(req.Username)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	if user == nil || !user.IsActive {
		response.Unauthorized(c, "invalid username or password")
		return
	}

	if !h.userSvc.VerifyPassword(req.Password, user.Password) {
		response.Unauthorized(c, "invalid username or password")
		return
	}

	token, err := auth.GenerateToken(uint(user.ID), user.Username, user.Role, h.jwtSecret, 24)
	if err != nil {
		response.InternalError(c, "failed to generate token")
		return
	}

	response.OK(c, gin.H{
		"token": token,
		"user": gin.H{
			"id":         user.ID,
			"username":   user.Username,
			"full_name":  user.FullName,
			"email":      user.Email,
			"title":      user.Title,
			"company":    user.Company,
			"role":       user.Role,
			"no_emp":     user.NoEmp,
			"id_emp":     user.IdEmp,
			"department": user.Department,
			"section":    user.Section,
			"division":   user.Division,
		},
	}, "login successful")
}

func (h *AuthHandler) GetMe(c *gin.Context) {
	userIDVal, _ := c.Get("user_id")
	var userID uint64
	switch v := userIDVal.(type) {
	case uint:
		userID = uint64(v)
	case uint64:
		userID = v
	case int:
		userID = uint64(v)
	}

	user, err := h.userSvc.GetByID(userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	if user == nil {
		response.NotFound(c, "user not found")
		return
	}

	response.OK(c, gin.H{
		"id":         user.ID,
		"username":   user.Username,
		"full_name":  user.FullName,
		"email":      user.Email,
		"title":      user.Title,
		"company":    user.Company,
		"role":       user.Role,
		"no_emp":     user.NoEmp,
		"id_emp":     user.IdEmp,
		"department": user.Department,
		"section":    user.Section,
		"division":   user.Division,
	}, "success")
}
