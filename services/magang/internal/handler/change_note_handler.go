package handler

import (
	"strconv"
	"strings"

	"magang-be/services/magang/internal/models"
	"magang-be/services/magang/internal/service"
	"magang-be/services/magang/pkg/response"

	"github.com/gin-gonic/gin"
)

type ChangeNoteHandler struct {
	svc     *service.ChangeNoteService
	userSvc *service.UserService
}

func NewChangeNoteHandler(svc *service.ChangeNoteService, userSvc *service.UserService) *ChangeNoteHandler {
	return &ChangeNoteHandler{svc: svc, userSvc: userSvc}
}

type CreateChangeNoteRequest struct {
	Initiator        string `json:"initiator"`
	EmpIDInitiator   string `json:"emp_id_initiator"`
	Department       string `json:"department"`
	ChangePertainsTo string `json:"change_pertains_to"`
}

func (h *ChangeNoteHandler) Create(c *gin.Context) {
	var req CreateChangeNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if strings.TrimSpace(req.ChangePertainsTo) == "" {
		response.BadRequest(c, "change_pertains_to is required")
		return
	}

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

	// Auto-populate from user profile if not provided
	initiator := strings.TrimSpace(req.Initiator)
	empID := strings.TrimSpace(req.EmpIDInitiator)
	dept := strings.TrimSpace(req.Department)

	if userID > 0 {
		user, _ := h.userSvc.GetByID(userID)
		if user != nil {
			if initiator == "" {
				initiator = user.FullName
			}
			if empID == "" {
				empID = user.IdEmp
			}
			if dept == "" {
				dept = user.Department
			}
		}
	}

	cn := &models.ChangeNote{
		UserID:           userID,
		Initiator:        initiator,
		EmpIDInitiator:   empID,
		Department:       dept,
		ChangePertainsTo: strings.TrimSpace(req.ChangePertainsTo),
		Status:           "SUBMITTED",
	}

	if err := h.svc.Create(cn); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Created(c, cn, "change note submitted successfully")
}

func (h *ChangeNoteHandler) GetAll(c *gin.Context) {
	search := strings.TrimSpace(c.Query("search"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

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

	roleVal, _ := c.Get("role")
	role, _ := roleVal.(string)

	changeNotes, total, err := h.svc.GetAll(search, userID, role, page, limit)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, gin.H{
		"data": changeNotes,
		"pagination": gin.H{
			"page":     page,
			"limit":    limit,
			"total":    total,
			"has_next": int64(page*limit) < total,
		},
	}, "success")
}

func (h *ChangeNoteHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id format")
		return
	}

	cn, err := h.svc.GetByID(id)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	if cn == nil {
		response.NotFound(c, "change note not found")
		return
	}

	response.OK(c, cn, "success")
}
