package repo

import (
	"context"

	"github.com/go-ddd-seed/go-ddd-seed/internal/domain/model/permission"
)

// PermissionRepository 权限仓储接口（命令侧）
type PermissionRepository interface {
	// Save 保存权限
	Save(ctx context.Context, permission *permission.Permission) error
	// FindByID 按 ID 查询
	FindByID(ctx context.Context, id string) (*permission.Permission, error)
	// FindByCode 按 Code 查询，用于唯一性校验
	FindByCode(ctx context.Context, code string) (*permission.Permission, error)
	// Delete 删除权限
	Delete(ctx context.Context, id string) error
}