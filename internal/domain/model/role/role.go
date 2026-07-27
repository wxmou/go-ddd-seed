package role

import (
	"time"

	"github.com/google/uuid"
	"github.com/go-ddd-seed/go-ddd-seed/internal/domain/model"
)

// RoleStatus 角色状态
const (
	RoleStatusDisabled = 0
	RoleStatusEnabled  = 1
)

// Role 角色聚合根
// 包含 RolePermission 子实体列表
type Role struct {
	model.AggregateRoot
	ID          string
	Name        string
	Code        string
	Description string
	Status      int
	Permissions []*RolePermission
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewRole 创建新角色
func NewRole(id, name, code, description string) (*Role, error) {
	if name == "" {
		return nil, ErrRoleNameEmpty
	}
	if code == "" {
		return nil, ErrRoleCodeEmpty
	}
	if len(code) > 100 {
		return nil, ErrRoleCodeTooLong
	}

	now := time.Now()
	return &Role{
		ID:          id,
		Name:        name,
		Code:        code,
		Description: description,
		Status:      RoleStatusEnabled,
		Permissions: make([]*RolePermission, 0),
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// Update 更新角色信息
func (r *Role) Update(name, description string) {
	r.Name = name
	r.Description = description
	r.UpdatedAt = time.Now()
}

// Enable 启用角色
func (r *Role) Enable() error {
	if r.Status == RoleStatusEnabled {
		return ErrRoleAlreadyEnabled
	}
	r.Status = RoleStatusEnabled
	r.UpdatedAt = time.Now()
	return nil
}

// Disable 禁用角色
func (r *Role) Disable() error {
	if r.Status == RoleStatusDisabled {
		return ErrRoleAlreadyDisabled
	}
	r.Status = RoleStatusDisabled
	r.UpdatedAt = time.Now()
	return nil
}

// IsEnabled 是否启用
func (r *Role) IsEnabled() bool {
	return r.Status == RoleStatusEnabled
}

// AssignPermission 分配权限
func (r *Role) AssignPermission(permissionID string) {
	for _, rp := range r.Permissions {
		if rp.PermissionID == permissionID {
			return
		}
	}
	r.Permissions = append(r.Permissions, &RolePermission{
		ID:           uuid.New().String(),
		RoleID:       r.ID,
		PermissionID: permissionID,
		CreatedAt:    time.Now(),
	})
	r.UpdatedAt = time.Now()
}

// RemovePermission 移除权限
func (r *Role) RemovePermission(permissionID string) {
	for i, rp := range r.Permissions {
		if rp.PermissionID == permissionID {
			r.Permissions = append(r.Permissions[:i], r.Permissions[i+1:]...)
			r.UpdatedAt = time.Now()
			return
		}
	}
}

// GetPermissionIDs 获取所有权限 ID
func (r *Role) GetPermissionIDs() []string {
	ids := make([]string, 0, len(r.Permissions))
	for _, rp := range r.Permissions {
		ids = append(ids, rp.PermissionID)
	}
	return ids
}