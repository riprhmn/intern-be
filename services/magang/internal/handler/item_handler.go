package handler

import (
	"strconv"
	"strings"

	"magang-be/services/magang/internal/models"
	"magang-be/services/magang/internal/service"
	"magang-be/services/magang/pkg/response"

	"github.com/gin-gonic/gin"
)

type ItemHandler struct {
	svc *service.ItemService
}

func NewItemHandler(svc *service.ItemService) *ItemHandler {
	return &ItemHandler{svc: svc}
}

func (h *ItemHandler) GetAll(c *gin.Context) {
	search := strings.TrimSpace(c.Query("search"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "25"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 25
	}

	items, total, err := h.svc.GetAll(search, page, limit)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, gin.H{
		"data": items,
		"pagination": gin.H{
			"page":     page,
			"limit":    limit,
			"total":    total,
			"has_next": int64(page*limit) < total,
		},
	}, "success")
}

func (h *ItemHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id format")
		return
	}

	item, err := h.svc.GetByID(id)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	if item == nil {
		response.NotFound(c, "item not found")
		return
	}

	response.OK(c, item, "success")
}

type CreateItemRequest struct {
	Name string `json:"name"`
	Note string `json:"note"`
}

func (h *ItemHandler) Create(c *gin.Context) {
	var req CreateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		response.BadRequest(c, "name is required")
		return
	}

	item := &models.Item{
		Name:   strings.TrimSpace(req.Name),
		Status: models.ItemStatusActive,
	}
	if req.Note != "" {
		item.Note = &req.Note
	}

	if err := h.svc.Create(item); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Created(c, item, "item created")
}

type UpdateItemRequest struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Note   string `json:"note"`
}

func (h *ItemHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id format")
		return
	}

	existing, err := h.svc.GetByID(id)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	if existing == nil {
		response.NotFound(c, "item not found")
		return
	}

	var req UpdateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.Status != "" {
		existing.Status = models.ItemStatus(req.Status)
	}
	if req.Note != "" {
		existing.Note = &req.Note
	}

	if err := h.svc.Update(existing); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, existing, "item updated")
}

func (h *ItemHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id format")
		return
	}

	if err := h.svc.Delete(id); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, nil, "item deleted")
}
