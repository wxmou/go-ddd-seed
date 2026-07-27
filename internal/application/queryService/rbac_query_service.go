package queryService

import (
	"context"

	"github.com/go-ddd-seed/go-ddd-seed/internal/application/port/repo"
)

// RoleQueryService 角色查询服务
type RoleQueryService struct {
	roleRead repo.RoleReadRepository
}

// NewRoleQueryService 创建角色查询服务
func NewRoleQueryService(roleRead repo.RoleReadRepository) *RoleQueryService {
	return &RoleQueryService{roleRead: roleRead}
}

// GetByID 按 ID 查询角色
func (s *RoleQueryService) GetByID(ctx context.Context, id string) (*repo.RoleDTO, error) {
	return s.roleRead.FindByID(ctx, id)
}

// List 角色列表
func (s *RoleQueryService) List(ctx context.Context) ([]*repo.RoleDTO, error) {
	return s.roleRead.List(ctx)
}

// PermissionQueryService 权限查询服务
type PermissionQueryService struct {
	permRead repo.PermissionReadRepository
}

// NewPermissionQueryService 创建权限查询服务
func NewPermissionQueryService(permRead repo.PermissionReadRepository) *PermissionQueryService {
	return &PermissionQueryService{permRead: permRead}
}

// List 权限列表
func (s *PermissionQueryService) List(ctx context.Context) ([]*repo.PermissionDTO, error) {
	return s.permRead.List(ctx)
}

// UserQueryService 用户查询服务
type UserQueryService struct {
	userRead repo.UserReadRepository
}

// NewUserQueryService 创建用户查询服务
func NewUserQueryService(userRead repo.UserReadRepository) *UserQueryService {
	return &UserQueryService{userRead: userRead}
}

// List 用户列表
func (s *UserQueryService) List(ctx context.Context, offset, limit, status int) (*PaginatedDTO, error) {
	list, total, err := s.userRead.List(ctx, offset, limit, status)
	if err != nil {
		return nil, err
	}
	return &PaginatedDTO{
		List:  list,
		Total: total,
		Page:  offset/limit + 1,
		Size:  limit,
	}, nil
}

// GetByID 按 ID 查询用户
func (s *UserQueryService) GetByID(ctx context.Context, id string) (*repo.UserDTO, error) {
	return s.userRead.FindByID(ctx, id)
}

// GetByIDWithRoles 按 ID 查询用户含角色权限
func (s *UserQueryService) GetByIDWithRoles(ctx context.Context, id string) (*repo.UserWithRolesDTO, error) {
	return s.userRead.FindByIDWithRoles(ctx, id)
}