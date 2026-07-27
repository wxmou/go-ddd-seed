package permission

import (
	"time"

	"github.com/go-ddd-seed/go-ddd-seed/internal/domain"
	"github.com/go-ddd-seed/go-ddd-seed/internal/domain/model"
)

// Permission 权限聚合根
type Permission struct {
	model.AggregateRoot
	ID          string
	Name        string
	Code        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewPermission 创建新权限
func NewPermission(id, name, code, description string) (*Permission, error) {
	if name == "" {
		return nil, ErrPermissionNameEmpty
	}
	if code == "" {
		return nil, ErrPermissionCodeEmpty
	}
	if len(code) > 200 {
		return nil, ErrPermissionCodeTooLong
	}

	now := time.Now()
	return &Permission{
		ID:          id,
		Name:        name,
		Code:        code,
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// Update 更新权限信息
func (p *Permission) Update(name, description string) {
	p.Name = name
	p.Description = description
	p.UpdatedAt = time.Now()
}

// Permission 聚合错误码范围: 3200-3299
var (
	ErrPermissionNameEmpty   = domain.NewDomainError(3200, "权限名称不能为空")
	ErrPermissionCodeEmpty   = domain.NewDomainError(3201, "权限编码不能为空")
	ErrPermissionCodeTooLong = domain.NewDomainError(3202, "权限编码长度不能超过200")
)