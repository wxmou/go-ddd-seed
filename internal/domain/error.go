package domain

// DomainError 领域层错误（无 HTTP 概念）
// 领域层禁止依赖 pkg/errors（含 HTTPCode），此类型是领域层唯一的错误载体
type DomainError struct {
	Code    int
	Message string
}

// Error 实现 error 接口
func (e *DomainError) Error() string {
	return e.Message
}

// NewDomainError 创建领域错误
func NewDomainError(code int, message string) *DomainError {
	return &DomainError{Code: code, Message: message}
}

// ErrRecordNotFound 资源不存在（领域层通用错误，仓储实现使用此错误而非 pkg/errors.ErrNotFound）
var ErrRecordNotFound = NewDomainError(404, "资源不存在")