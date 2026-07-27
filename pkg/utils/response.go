package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-ddd-seed/go-ddd-seed/internal/domain"
	appErrors "github.com/go-ddd-seed/go-ddd-seed/pkg/errors"
	"github.com/go-ddd-seed/go-ddd-seed/pkg/validator"
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "ok",
		Data:    data,
	})
}

// SuccessWithMsg 带消息的成功响应
func SuccessWithMsg(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: message,
		Data:    data,
	})
}

// Fail 失败响应
func Fail(c *gin.Context, err *appErrors.Error) {
	c.JSON(err.HTTPCode, Response{
		Code:    err.Code,
		Message: err.Message,
	})
}

// FailWithMsg 自定义失败响应
func FailWithMsg(c *gin.Context, httpCode int, code int, message string) {
	c.JSON(httpCode, Response{
		Code:    code,
		Message: message,
	})
}

// ValidationFail 校验失败响应（统一字段级错误格式）
func ValidationFail(c *gin.Context, err *validator.ValidateError) {
	c.JSON(http.StatusBadRequest, Response{
		Code:    err.Code,
		Message: err.Message,
		Data:    err.Fields,
	})
}

// Error 错误响应（用于未知错误）
func Error(c *gin.Context, err error) {
	if appErr, ok := err.(*appErrors.Error); ok {
		Fail(c, appErr)
		return
	}
	if domErr, ok := err.(*domain.DomainError); ok {
		c.JSON(http.StatusBadRequest, Response{
			Code:    domErr.Code,
			Message: domErr.Message,
		})
		return
	}
	Fail(c, appErrors.ErrInternal)
}