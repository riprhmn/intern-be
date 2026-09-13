package handler

import (
	"magang-be/services/magang/internal/service"
	"magang-be/services/magang/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ANCRHandler struct {
	svc     *service.ANCRService
	userSvc *service.UserService
}

func NewANCRHandler(svc *service.ANCRService, userSvc *service.UserService) *ANCRHandler {
	return &ANCRHandler{svc: svc, userSvc: userSvc}
}

func (h *ANCRHandler) ancrActor(c *gin.Context) uint64 {
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
	return id
}

// GET /api/v1/magang/ancr
func (h *ANCRHandler) GetAll(c *gin.Context) {
	p, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	l, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	if p < 1 {
		p = 1
	}
	if l < 1 || l > 500 {
		l = 100
	}
	search := c.Query("search")
	rows, total, err := h.svc.GetAll(search, p, l)
	if err != nil {
		response.InternalError(c, "Gagal memuat data ANCR")
		return
	}
	response.OK(c, gin.H{
		"data":       rows,
		"pagination": gin.H{"page": p, "limit": l, "total": total, "has_next": int64(p*l) < total},
	}, "success")
}

// GET /api/v1/magang/ancr/:id
func (h *ANCRHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID tidak valid")
		return
	}
	row, err := h.svc.GetByID(id)
	if err != nil {
		response.NotFound(c, "ANCR tidak ditemukan")
		return
	}
	response.OK(c, row, "success")
}

// POST /api/v1/magang/ancr
func (h *ANCRHandler) Create(c *gin.Context) {
	userID := h.ancrActor(c)
	var req service.ANCRCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Data tidak valid: "+err.Error())
		return
	}
	row, err := h.svc.Create(userID, req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, row, "ANCR berhasil dibuat")
}

// PUT /api/v1/magang/ancr/:id
func (h *ANCRHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID tidak valid")
		return
	}
	var req service.ANCRUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Data tidak valid")
		return
	}
	row, err := h.svc.Update(id, req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, row, "ANCR diperbarui")
}

// DELETE /api/v1/magang/ancr/:id
func (h *ANCRHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID tidak valid")
		return
	}
	if err := h.svc.Delete(id); err != nil {
		response.InternalError(c, "Gagal menghapus ANCR")
		return
	}
	response.OK(c, nil, "ANCR dihapus")
}
