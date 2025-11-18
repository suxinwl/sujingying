package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// SuccessResponse 成功响应的标准格式
type SuccessResponse struct {
	Data    interface{} `json:"data"`
	Message string      `json:"message,omitempty"`
}

// ErrorResponse 错误响应的标准格式
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// Success 返回成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, SuccessResponse{
		Data: data,
	})
}

// SuccessWithMessage 返回带消息的成功响应
func SuccessWithMessage(c *gin.Context, data interface{}, message string) {
	c.JSON(http.StatusOK, SuccessResponse{
		Data:    data,
		Message: message,
	})
}

// Error 返回错误响应
func Error(c *gin.Context, statusCode int, err string) {
	c.JSON(statusCode, ErrorResponse{
		Error: err,
	})
}

// BadRequest 返回400错误
func BadRequest(c *gin.Context, err string) {
	Error(c, http.StatusBadRequest, err)
}

// Unauthorized 返回401错误
func Unauthorized(c *gin.Context, err string) {
	Error(c, http.StatusUnauthorized, err)
}

// Forbidden 返回403错误
func Forbidden(c *gin.Context, err string) {
	Error(c, http.StatusForbidden, err)
}

// NotFound 返回404错误
func NotFound(c *gin.Context, err string) {
	Error(c, http.StatusNotFound, err)
}

// InternalServerError 返回500错误
func InternalServerError(c *gin.Context, err string) {
	Error(c, http.StatusInternalServerError, err)
}
