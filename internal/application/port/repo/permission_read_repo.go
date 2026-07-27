package repo

import (
	"context"
	"time"
)

// PermissionDTO 权限读模型 DTO
type PermissionDTO struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// PermissionReadRepository 权限读仓储接口
type PermissionReadRepository interface {
	// FindByCode 按 Code 查询权限
	FindByCode(ctx context.Context, code string) (*PermissionDTO, error)
	// List 权限列表
	List(ctx context.Context) ([]*PermissionDTO, error)
}