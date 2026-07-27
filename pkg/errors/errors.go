package errors

import "net/http"

// Error 应用通用错误
type Error struct {
	Code     int    `json:"code"`
	Message  string `json:"message"`
	HTTPCode int    `json:"-"`
}

// Error 实现 error 接口
func (e *Error) Error() string {
	return e.Message
}

// New 创建通用错误
func New(code int, message string) *Error {
	return &Error{Code: code, Message: message, HTTPCode: http.StatusBadRequest}
}

// NewWithHTTP 创建带 HTTP 状态码的错误
func NewWithHTTP(code int, message string, httpCode int) *Error {
	return &Error{Code: code, Message: message, HTTPCode: httpCode}
}

// 预定义错误码
var (
	// 通用错误 (0-999)
	ErrSuccess      = &Error{Code: 0, Message: "ok", HTTPCode: http.StatusOK}
	ErrUnknown      = NewWithHTTP(1, "未知错误", http.StatusInternalServerError)
	ErrBadRequest   = New(2, "请求参数错误")
	ErrUnauthorized = NewWithHTTP(3, "未授权", http.StatusUnauthorized)
	ErrForbidden    = NewWithHTTP(4, "无权限", http.StatusForbidden)
	ErrNotFound     = NewWithHTTP(5, "资源不存在", http.StatusNotFound)
	ErrConflict     = New(6, "资源冲突")
	ErrInternal     = NewWithHTTP(7, "服务器内部错误", http.StatusInternalServerError)

	// 业务错误 (1000-1999)
	ErrInvalidCredentials = NewWithHTTP(1000, "用户名或密码错误", http.StatusUnauthorized)
	ErrTokenExpired       = NewWithHTTP(1001, "令牌已过期", http.StatusUnauthorized)
	ErrTokenInvalid       = NewWithHTTP(1002, "无效的令牌", http.StatusUnauthorized)
	ErrUserDisabled       = NewWithHTTP(1003, "用户已被禁用", http.StatusForbidden)

	// 领域错误 (2000-2999)
	ErrInvalidStatus = New(2000, "无效的状态转换")
	ErrBusinessRule  = New(2001, "违反业务规则")
)
