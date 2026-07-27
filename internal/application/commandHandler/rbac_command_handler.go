package commandHandler

import (
	"context"

	"github.com/google/uuid"
	appRepo "github.com/go-ddd-seed/go-ddd-seed/internal/application/port/repo"
	"github.com/go-ddd-seed/go-ddd-seed/internal/application/command"
	domainRepo "github.com/go-ddd-seed/go-ddd-seed/internal/domain/port/repo"
	"github.com/go-ddd-seed/go-ddd-seed/internal/domain/model/permission"
	"github.com/go-ddd-seed/go-ddd-seed/internal/domain/model/role"
)

// RbacCommandHandler RBAC 命令处理器
type RbacCommandHandler struct {
	roleRepo domainRepo.RoleRepository
	permRepo domainRepo.PermissionRepository
	userRepo domainRepo.UserRepository
}

// NewRbacCommandHandler 创建 RBAC 命令处理器
func NewRbacCommandHandler(
	roleRepo domainRepo.RoleRepository,
	permRepo domainRepo.PermissionRepository,
	userRepo domainRepo.UserRepository,
) *RbacCommandHandler {
	return &RbacCommandHandler{
		roleRepo: roleRepo,
		permRepo: permRepo,
		userRepo: userRepo,
	}
}

// ----- 角色管理 -----

// CreateRole 创建角色
func (h *RbacCommandHandler) CreateRole(ctx context.Context, cmd *command.CreateRoleCommand) (*appRepo.RoleDTO, error) {
	// 检查 code 唯一性
	existing, _ := h.roleRepo.FindByCode(ctx, cmd.Code)
	if existing != nil {
		return nil, ErrRoleCodeDuplicate
	}

	role, err := role.NewRole(uuid.New().String(), cmd.Name, cmd.Code, cmd.Description)
	if err != nil {
		return nil, err
	}

	if err := h.roleRepo.Save(ctx, role); err != nil {
		return nil, err
	}

	return &appRepo.RoleDTO{
		ID:          role.ID,
		Name:        role.Name,
		Code:        role.Code,
		Description: role.Description,
		Status:      role.Status,
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	}, nil
}

// UpdateRole 更新角色
func (h *RbacCommandHandler) UpdateRole(ctx context.Context, cmd *command.UpdateRoleCommand) (*appRepo.RoleDTO, error) {
	role, err := h.roleRepo.FindByID(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}

	role.Update(cmd.Name, cmd.Description)

	if err := h.roleRepo.Save(ctx, role); err != nil {
		return nil, err
	}

	return &appRepo.RoleDTO{
		ID:          role.ID,
		Name:        role.Name,
		Code:        role.Code,
		Description: role.Description,
		Status:      role.Status,
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	}, nil
}

// DeleteRole 删除角色
func (h *RbacCommandHandler) DeleteRole(ctx context.Context, cmd *command.DeleteRoleCommand) error {
	return h.roleRepo.Delete(ctx, cmd.ID)
}

// EnableRole 启用角色
func (h *RbacCommandHandler) EnableRole(ctx context.Context, cmd *command.EnableRoleCommand) (*appRepo.RoleWithPermissionsDTO, error) {
	role, err := h.roleRepo.FindByID(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}

	if err := role.Enable(); err != nil {
		return nil, err
	}

	if err := h.roleRepo.Save(ctx, role); err != nil {
		return nil, err
	}

	return h.toRoleWithPermsDTO(role)
}

// DisableRole 禁用角色
func (h *RbacCommandHandler) DisableRole(ctx context.Context, cmd *command.DisableRoleCommand) (*appRepo.RoleWithPermissionsDTO, error) {
	role, err := h.roleRepo.FindByID(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}

	if err := role.Disable(); err != nil {
		return nil, err
	}

	if err := h.roleRepo.Save(ctx, role); err != nil {
		return nil, err
	}

	return h.toRoleWithPermsDTO(role)
}

// ----- 权限管理 -----

// CreatePermission 创建权限
func (h *RbacCommandHandler) CreatePermission(ctx context.Context, cmd *command.CreatePermissionCommand) (*appRepo.PermissionDTO, error) {
	// 检查 code 唯一性
	existing, _ := h.permRepo.FindByCode(ctx, cmd.Code)
	if existing != nil {
		return nil, ErrPermissionCodeDuplicate
	}

	perm, err := permission.NewPermission(uuid.New().String(), cmd.Name, cmd.Code, cmd.Description)
	if err != nil {
		return nil, err
	}

	if err := h.permRepo.Save(ctx, perm); err != nil {
		return nil, err
	}

	return &appRepo.PermissionDTO{
		ID:          perm.ID,
		Name:        perm.Name,
		Code:        perm.Code,
		Description: perm.Description,
		CreatedAt:   perm.CreatedAt,
		UpdatedAt:   perm.UpdatedAt,
	}, nil
}

// UpdatePermission 更新权限
func (h *RbacCommandHandler) UpdatePermission(ctx context.Context, cmd *command.UpdatePermissionCommand) (*appRepo.PermissionDTO, error) {
	perm, err := h.permRepo.FindByID(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}

	perm.Update(cmd.Name, cmd.Description)

	if err := h.permRepo.Save(ctx, perm); err != nil {
		return nil, err
	}

	return &appRepo.PermissionDTO{
		ID:          perm.ID,
		Name:        perm.Name,
		Code:        perm.Code,
		Description: perm.Description,
		CreatedAt:   perm.CreatedAt,
		UpdatedAt:   perm.UpdatedAt,
	}, nil
}

// DeletePermission 删除权限
func (h *RbacCommandHandler) DeletePermission(ctx context.Context, cmd *command.DeletePermissionCommand) error {
	return h.permRepo.Delete(ctx, cmd.ID)
}

// ----- 角色-权限绑定 -----

// AssignPermissionToRole 为角色分配权限
func (h *RbacCommandHandler) AssignPermissionToRole(ctx context.Context, cmd *command.AssignRolePermissionCommand) (*appRepo.RoleWithPermissionsDTO, error) {
	role, err := h.roleRepo.FindByID(ctx, cmd.RoleID)
	if err != nil {
		return nil, err
	}

	if cmd.PermissionIDs != nil {
		for _, permID := range cmd.PermissionIDs {
			role.AssignPermission(permID)
		}
	} else if cmd.PermissionID != "" {
		role.AssignPermission(cmd.PermissionID)
	}

	if err := h.roleRepo.Save(ctx, role); err != nil {
		return nil, err
	}

	return h.toRoleWithPermsDTO(role)
}

// RemovePermissionFromRole 移除角色权限
func (h *RbacCommandHandler) RemovePermissionFromRole(ctx context.Context, cmd *command.RemoveRolePermissionCommand) (*appRepo.RoleWithPermissionsDTO, error) {
	role, err := h.roleRepo.FindByID(ctx, cmd.RoleID)
	if err != nil {
		return nil, err
	}

	role.RemovePermission(cmd.PermissionID)

	if err := h.roleRepo.Save(ctx, role); err != nil {
		return nil, err
	}

	return h.toRoleWithPermsDTO(role)
}

// ----- 用户-角色绑定 -----

// AssignRoleToUser 为用户分配角色
func (h *RbacCommandHandler) AssignRoleToUser(ctx context.Context, cmd *command.AssignUserRoleCommand) error {
	user, err := h.userRepo.FindByID(ctx, cmd.UserID)
	if err != nil {
		return err
	}

	user.AssignRole(cmd.RoleID)

	return h.userRepo.Save(ctx, user)
}

// RemoveRoleFromUser 移除用户角色
func (h *RbacCommandHandler) RemoveRoleFromUser(ctx context.Context, cmd *command.RemoveUserRoleCommand) error {
	user, err := h.userRepo.FindByID(ctx, cmd.UserID)
	if err != nil {
		return err
	}

	user.RemoveRole(cmd.RoleID)

	return h.userRepo.Save(ctx, user)
}

// ----- helpers -----

func (h *RbacCommandHandler) toRoleWithPermsDTO(role *role.Role) (*appRepo.RoleWithPermissionsDTO, error) {
	permDTOs := make([]*appRepo.PermissionDTO, 0, len(role.Permissions))
	for _, rp := range role.Permissions {
		permDTOs = append(permDTOs, &appRepo.PermissionDTO{
			ID: rp.PermissionID,
		})
	}

	return &appRepo.RoleWithPermissionsDTO{
		RoleDTO: appRepo.RoleDTO{
			ID:          role.ID,
			Name:        role.Name,
			Code:        role.Code,
			Description: role.Description,
			Status:      role.Status,
			CreatedAt:   role.CreatedAt,
			UpdatedAt:   role.UpdatedAt,
		},
		Permissions: permDTOs,
	}, nil
}