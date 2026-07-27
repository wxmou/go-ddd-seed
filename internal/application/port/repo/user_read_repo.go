package repo

import (
	"context"
	"time"
)

// UserDTO 用户读模型 DTO（CQRS 查询专用）
type UserDTO struct {
	ID          string     `json:"id"`
	Username    string     `json:"username"`
	RealName    string     `json:"real_name"`
	Email       string     `json:"email"`
	Phone       string     `json:"phone"`
	Status      int        `json:"status"`
	LastLoginAt *time.Time `json:"last_login_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// RoleEntry 角色条目（用户信息中的角色）
type RoleEntry struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

// PermissionEntry 权限条目
type PermissionEntry struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

// UserWithRolesDTO 用户含角色权限的 DTO
type UserWithRolesDTO struct {
	UserDTO
	Roles       []*RoleEntry       `json:"roles"`
	Permissions []*PermissionEntry `json:"permissions"`
}

// UserReadRepository 用户读仓储接口（CQRS 读模型）
type UserReadRepository interface {
	// FindByID 按 ID 查询用户基本信息
	FindByID(ctx context.Context, id string) (*UserDTO, error)
	// FindByUsername 按用户名查询用户（含角色+权限）
	FindByUsername(ctx context.Context, username string) (*UserWithRolesDTO, error)
	// FindByIDWithRoles 按 ID 查询用户含角色权限
	FindByIDWithRoles(ctx context.Context, id string) (*UserWithRolesDTO, error)
	// List 用户列表（分页+状态筛选, status=-1 表示全部）
	List(ctx context.Context, offset, limit int, status int) ([]*UserDTO, int64, error)
}