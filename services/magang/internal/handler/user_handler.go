package handler

import (
	"strconv"
	"strings"

	"magang-be/services/magang/internal/repository"
	"magang-be/services/magang/internal/service"
	"magang-be/services/magang/pkg/response"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	svc *service.UserService
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

func (h *UserHandler) GetAll(c *gin.Context) {
	search := strings.TrimSpace(c.Query("search"))
	role := strings.TrimSpace(c.Query("role"))
	section := strings.TrimSpace(c.Query("section"))
	department := strings.TrimSpace(c.Query("department"))
	division := strings.TrimSpace(c.Query("division"))

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	filter := repository.UserFilter{
		Search:     search,
		Role:       role,
		Section:    section,
		Department: department,
		Division:   division,
		Page:       page,
		Limit:      limit,
	}

	users, total, err := h.svc.GetAll(filter)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, gin.H{
		"data": users,
		"pagination": gin.H{
			"page":     page,
			"limit":    limit,
			"total":    total,
			"has_next": int64(page*limit) < total,
		},
	}, "success")
}

func (h *UserHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id format")
		return
	}

	user, err := h.svc.GetByID(id)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	if user == nil {
		response.NotFound(c, "user not found")
		return
	}

	response.OK(c, user, "success")
}

func (h *UserHandler) Create(c *gin.Context) {
	var req service.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.svc.Create(req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Created(c, nil, "user created successfully")
}

func (h *UserHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id format")
		return
	}

	var req service.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.svc.Update(id, req); err != nil {
		if err.Error() == "user not found" {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, nil, "user updated successfully")
}

func (h *UserHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id format")
		return
	}

	if err := h.svc.Delete(id); err != nil {
		if err.Error() == "user not found" {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, nil, "user deleted successfully")
}
