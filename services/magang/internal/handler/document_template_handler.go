package handler

import (
	"github.com/gin-gonic/gin"
	"magang-be/services/magang/internal/models"
	"magang-be/services/magang/internal/service"
	"magang-be/services/magang/pkg/response"
	"strconv"
	"strings"
)

func (h *ChangeNoteHandler) Templates(c *gin.Context) {
	u := h.actor(c)
	if u == nil {
		return
	}
	id, _ := strconv.ParseUint(c.Param("templateID"), 10, 64)
	switch c.Request.Method {
	case "GET":
		if id == 0 {
			rows, err := h.svc.Templates()
			if err != nil {
				response.InternalError(c, "Gagal memuat template")
				return
			}
			response.OK(c, rows, "success")
		} else {
			row, err := h.svc.Template(id)
			if err != nil {
				response.NotFound(c, "Template tidak ditemukan")
				return
			}
			response.OK(c, row, "success")
		}
	case "PUT":
		if u.Role != "admin" {
			response.Forbidden(c, "Hanya admin dapat mengelola template")
			return
		}
		var row models.DocumentTemplate
		if err := c.ShouldBindJSON(&row); err != nil {
			response.BadRequest(c, "Template tidak valid")
			return
		}
		row.ID = id
		if strings.TrimSpace(row.Title) == "" || strings.TrimSpace(row.Category) == "" || strings.TrimSpace(row.Version) == "" {
			response.BadRequest(c, "Judul, kategori dan versi wajib diisi")
			return
		}
		if err := service.ValidateCNDocument(row.FileName, row.FileData); err != nil {
			response.BadRequest(c, err.Error())
			return
		}
		if err := h.svc.SaveTemplate(&row); err != nil {
			response.BadRequest(c, "Gagal menyimpan template")
			return
		}
		response.OK(c, row, "Template tersimpan")
	case "DELETE":
		if u.Role != "admin" {
			response.Forbidden(c, "Hanya admin dapat menghapus template")
			return
		}
		if id == 0 {
			response.BadRequest(c, "ID tidak valid")
			return
		}
		if err := h.svc.DeleteTemplate(id); err != nil {
			response.InternalError(c, "Gagal menghapus template")
			return
		}
		response.OK(c, nil, "Template dihapus")
	}
}
