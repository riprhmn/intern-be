package handler

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"magang-be/services/magang/internal/models"
	"magang-be/services/magang/internal/repository"
	"magang-be/services/magang/internal/service"
	"magang-be/services/magang/pkg/response"
	"strconv"
	"strings"
	"time"
)

type ChangeNoteHandler struct {
	svc     *service.ChangeNoteService
	userSvc *service.UserService
}

func NewChangeNoteHandler(s *service.ChangeNoteService, u *service.UserService) *ChangeNoteHandler {
	return &ChangeNoteHandler{svc: s, userSvc: u}
}
func (h *ChangeNoteHandler) actor(c *gin.Context) *models.User {
	v, _ := c.Get("user_id")
	var id uint64
	switch n := v.(type) {
	case uint64:
		id = n
	case uint:
		id = uint64(n)
	case int:
		id = uint64(n)
	}
	u, err := h.userSvc.GetByID(id)
	if err != nil || u == nil || !u.IsActive {
		response.Unauthorized(c, "Akun tidak aktif")
		return nil
	}
	return u
}
func cnError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrCNForbidden):
		response.Forbidden(c, err.Error())
	case errors.Is(err, service.ErrCNConflict):
		c.JSON(409, gin.H{"error": err.Error()})
	case errors.Is(err, gorm.ErrRecordNotFound):
		response.NotFound(c, "CN tidak ditemukan")
	default:
		response.BadRequest(c, err.Error())
	}
}
func (h *ChangeNoteHandler) Create(c *gin.Context) {
	u := h.actor(c)
	if u == nil {
		return
	}
	if u.Role == "external_audit" {
		response.Forbidden(c, "External audit hanya dapat membaca CN")
		return
	}
	var req struct {
		ChangePertainsTo string `json:"change_pertains_to"`
		DocumentName     string `json:"document_name"`
		DocumentData     string `json:"document_data"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Data request tidak valid")
		return
	}
	if strings.TrimSpace(req.ChangePertainsTo) == "" {
		response.BadRequest(c, "Keterangan perubahan wajib diisi")
		return
	}
	if err := service.ValidateCNDocument(req.DocumentName, req.DocumentData); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	cn := &models.ChangeNote{UserID: u.ID, Initiator: u.FullName, EmpIDInitiator: u.IdEmp, Department: u.Department, Section: u.Section, ChangePertainsTo: strings.TrimSpace(req.ChangePertainsTo), DocumentName: req.DocumentName, DocumentData: req.DocumentData, Status: "WAITING_ISO", Stages: []models.ApprovalStage{}, History: []models.CNEvent{{Action: "submit", ActorID: u.ID, ActorName: u.FullName, At: time.Now().UTC()}}}
	if err := h.svc.Create(cn); err != nil {
		response.InternalError(c, "Gagal menyimpan CN")
		return
	}
	response.Created(c, cn, "CN menunggu proses ISO")
}
func (h *ChangeNoteHandler) GetAll(c *gin.Context) {
	u := h.actor(c)
	if u == nil {
		return
	}
	p, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	l, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if p < 1 {
		p = 1
	}
	if l < 1 {
		l = 10
	}
	if l > 100 {
		l = 100
	}
	mode := c.Query("mode")
	notes, total, err := h.svc.Visible(u, strings.TrimSpace(c.Query("search")), mode, c.Query("department"), p, l)
	if err != nil {
		response.InternalError(c, "Gagal memuat CN")
		return
	}
	departments, err := h.svc.Departments(u, mode)
	if err != nil {
		response.InternalError(c, "Gagal memuat filter departemen")
		return
	}
	response.OK(c, gin.H{"data": notes, "filters": gin.H{"departments": departments}, "pagination": gin.H{"page": p, "limit": l, "total": total, "has_next": int64(p*l) < total}}, "success")
}
func (h *ChangeNoteHandler) GetByID(c *gin.Context) {
	u := h.actor(c)
	if u == nil {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID tidak valid")
		return
	}
	cn, err := h.svc.GetByID(id)
	if err != nil {
		response.InternalError(c, "Gagal memuat CN")
		return
	}
	if cn == nil {
		response.NotFound(c, "CN tidak ditemukan")
		return
	}
	if !service.CNCanRead(cn, u) {
		response.Forbidden(c, "Akses CN ditolak")
		return
	}
	response.OK(c, cn, "success")
}
func (h *ChangeNoteHandler) Update(c *gin.Context) {
	u := h.actor(c)
	if u == nil {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID tidak valid")
		return
	}
	var a service.CNAction
	if err = c.ShouldBindJSON(&a); err != nil {
		response.BadRequest(c, "Data tindakan tidak valid")
		return
	}
	cn, err := h.svc.Act(id, u, a)
	if err != nil {
		cnError(c, err)
		return
	}
	response.OK(c, cn, "CN diperbarui")
}
func (h *ChangeNoteHandler) Approvers(c *gin.Context) {
	if h.actor(c) == nil {
		return
	}
	p, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if p < 1 {
		p = 1
	}
	users, total, err := h.userSvc.GetAll(repository.UserFilter{Page: p, Limit: 100, Role: "approval"})
	if err != nil {
		response.InternalError(c, "Gagal memuat approver")
		return
	}
	data := []gin.H{}
	for _, u := range users {
		data = append(data, gin.H{"id": u.ID, "full_name": u.FullName, "username": u.Username, "department": u.Department, "title": u.Title, "section": u.Section, "division": u.Division, "role": u.Role, "is_active": u.IsActive})
	}
	response.OK(c, gin.H{"data": data, "pagination": gin.H{"has_next": int64(p*100) < total}}, "success")
}
func (h *ChangeNoteHandler) Summary(c *gin.Context) {
	u := h.actor(c)
	if u == nil {
		return
	}
	v, err := h.svc.Summary(u)
	if err != nil {
		response.InternalError(c, "Gagal memuat ringkasan CN")
		return
	}
	response.OK(c, v, "success")
}
