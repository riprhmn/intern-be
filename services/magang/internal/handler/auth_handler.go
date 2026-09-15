package handler

import (
	"log"
	"net/url"
	"os"
	"strings"

	"magang-be/services/magang/internal/service"
	"magang-be/services/magang/pkg/auth"
	"magang-be/services/magang/pkg/response"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	userSvc   *service.UserService
	emailSvc  *service.EmailService
	jwtSecret string
}

func NewAuthHandler(userSvc *service.UserService, emailSvc *service.EmailService, jwtSecret string) *AuthHandler {
	return &AuthHandler{userSvc: userSvc, emailSvc: emailSvc, jwtSecret: jwtSecret}
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

type ForgotPasswordRequest struct {
	Identity string `json:"identity"`
}

func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	req.Identity = strings.TrimSpace(req.Identity)
	if req.Identity == "" {
		response.BadRequest(c, "email or username is required")
		return
	}

	user, err := h.userSvc.GetByIdentity(req.Identity)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	if user == nil {
		response.BadRequest(c, "user not found")
		return
	}
	if strings.TrimSpace(user.Email) == "" {
		response.BadRequest(c, "email address is not registered for this account")
		return
	}

	token, err := h.userSvc.GenerateResetToken(req.Identity)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	frontendURL := strings.TrimRight(strings.TrimSpace(os.Getenv("FRONTEND_URL")), "/")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}
	resetLink := frontendURL + "/auth/reset-password?token=" + url.QueryEscape(token)
	if err := h.emailSvc.SendResetPasswordEmail(user.Email, resetLink); err != nil {
		log.Printf("failed to send reset password email: %v", err)
		response.InternalError(c, "failed to send reset password email")
		return
	}

	response.OK(c, nil, "reset password link has been sent to your email")
}

type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.userSvc.ResetPassword(req.Token, req.NewPassword); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, nil, "password reset successful")
}
