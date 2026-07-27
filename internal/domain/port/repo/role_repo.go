package repo

import (
	"context"

	"github.com/go-ddd-seed/go-ddd-seed/internal/domain/model/role"
)

// RoleRepository 角色仓储接口（命令侧）
type RoleRepository interface {
	// Save 保存角色（含权限关联）
	Save(ctx context.Context, role *role.Role) error
	// FindByID 按 ID 加载角色（含权限关联）
	FindByID(ctx context.Context, id string) (*role.Role, error)
	// FindByCode 按 Code 加载角色
	FindByCode(ctx context.Context, code string) (*role.Role, error)
	// Delete 删除角色
	Delete(ctx context.Context, id string) error
}