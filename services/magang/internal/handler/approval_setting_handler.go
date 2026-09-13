package handler

import (
	"errors"
	"strconv"

	"magang-be/services/magang/internal/service"
	"magang-be/services/magang/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ApprovalSettingHandler struct {
	svc *service.ApprovalSettingService
}

func NewApprovalSettingHandler(svc *service.ApprovalSettingService) *ApprovalSettingHandler {
	return &ApprovalSettingHandler{svc: svc}
}

func actorID(c *gin.Context) uint64 {
	value, _ := c.Get("user_id")
	switch id := value.(type) {
	case uint:
		return uint64(id)
	case uint64:
		return id
	case float64:
		return uint64(id)
	case int:
		return uint64(id)
	default:
		return 0
	}
}

func (h *ApprovalSettingHandler) Options(c *gin.Context) {
	data, err := h.svc.Options()
	if err != nil {
		response.InternalError(c, "Gagal memuat pilihan approval")
		return
	}
	response.OK(c, data, "success")
}

func (h *ApprovalSettingHandler) GetMapping(c *gin.Context) {
	rows, err := h.svc.Mapping(c.Query("approval_type"), c.Query("section_code"))
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, rows, "success")
}

func (h *ApprovalSettingHandler) ReplaceMapping(c *gin.Context) {
	var req service.CodeMappingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Data mapping tidak valid")
		return
	}
	row, err := h.svc.ReplaceMapping(c.Param("approvalType"), c.Param("sectionCode"), actorID(c), req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, row, "Mapping approval berhasil disimpan")
}

func (h *ApprovalSettingHandler) DeleteMapping(c *gin.Context) {
	err := h.svc.DeleteMapping(c.Param("approvalType"), c.Param("sectionCode"))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.NotFound(c, "Mapping tidak ditemukan")
		return
	}
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, nil, "Mapping dihapus; CN baru akan menggunakan workflow lama")
}

func (h *ApprovalSettingHandler) GetDelegations(c *gin.Context) {
	rows, err := h.svc.Delegations(c.Query("include_inactive") == "true")
	if err != nil {
		response.InternalError(c, "Gagal memuat delegasi")
		return
	}
	response.OK(c, rows, "success")
}

func (h *ApprovalSettingHandler) CreateDelegation(c *gin.Context) {
	var req service.DelegationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Data delegasi tidak valid")
		return
	}
	row, err := h.svc.CreateDelegation(actorID(c), req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, row, "Delegasi berhasil dibuat")
}

func (h *ApprovalSettingHandler) UpdateDelegation(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID delegasi tidak valid")
		return
	}
	var req service.DelegationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Data delegasi tidak valid")
		return
	}
	row, err := h.svc.UpdateDelegation(id, actorID(c), req)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.NotFound(c, "Delegasi tidak ditemukan")
		return
	}
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, row, "Delegasi berhasil diperbarui")
}

func (h *ApprovalSettingHandler) DeactivateDelegation(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID delegasi tidak valid")
		return
	}
	err = h.svc.DeactivateDelegation(id, actorID(c))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.NotFound(c, "Delegasi aktif tidak ditemukan")
		return
	}
	if err != nil {
		response.InternalError(c, "Gagal menonaktifkan delegasi")
		return
	}
	response.OK(c, nil, "Delegasi dinonaktifkan")
}
