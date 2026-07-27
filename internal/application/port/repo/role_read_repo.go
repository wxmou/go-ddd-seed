package repo

import (
	"context"
	"time"
)

// RoleDTO 角色读模型 DTO
type RoleDTO struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	Description string    `json:"description"`
	Status      int       `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// RoleWithPermissionsDTO 角色含权限的 DTO
type RoleWithPermissionsDTO struct {
	RoleDTO
	Permissions []*PermissionDTO `json:"permissions"`
}

// RoleReadRepository 角色读仓储接口
type RoleReadRepository interface {
	// FindByID 按 ID 查询角色
	FindByID(ctx context.Context, id string) (*RoleDTO, error)
	// FindByCode 按 Code 查询角色
	FindByCode(ctx context.Context, code string) (*RoleWithPermissionsDTO, error)
	// List 角色列表
	List(ctx context.Context) ([]*RoleDTO, error)
}