package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type envelope struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func OK(c *gin.Context, data interface{}, message string) {
	c.JSON(http.StatusOK, envelope{Success: true, Message: message, Data: data})
}

func Created(c *gin.Context, data interface{}, message string) {
	c.JSON(http.StatusCreated, envelope{Success: true, Message: message, Data: data})
}

func BadRequest(c *gin.Context, errMsg string) {
	c.JSON(http.StatusBadRequest, envelope{Success: false, Error: errMsg})
}

func Unauthorized(c *gin.Context, errMsg string) {
	c.JSON(http.StatusUnauthorized, envelope{Success: false, Error: errMsg})
}

func Forbidden(c *gin.Context, errMsg string) {
	c.JSON(http.StatusForbidden, envelope{Success: false, Error: errMsg})
}

func NotFound(c *gin.Context, errMsg string) {
	c.JSON(http.StatusNotFound, envelope{Success: false, Error: errMsg})
}

func InternalError(c *gin.Context, errMsg string) {
	c.JSON(http.StatusInternalServerError, envelope{Success: false, Error: errMsg})
}
