package repo

import (
	"context"

	"github.com/go-ddd-seed/go-ddd-seed/internal/domain"
	"gorm.io/gorm"

	appRepo "github.com/go-ddd-seed/go-ddd-seed/internal/application/port/repo"
)

// Ensure interface compliance
var _ appRepo.UserReadRepository = (*UserReadRepository)(nil)

// UserReadRepository 用户读仓储 GORM 实现
type UserReadRepository struct {
	db *gorm.DB
}

// NewUserReadRepository 创建用户读仓储
func NewUserReadRepository(db *gorm.DB) *UserReadRepository {
	return &UserReadRepository{db: db}
}

// FindByID 按 ID 查询用户基本信息
func (r *UserReadRepository) FindByID(ctx context.Context, id string) (*appRepo.UserDTO, error) {
	var gormUser UserGorm
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&gormUser).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrRecordNotFound
		}
		return nil, err
	}

	return &appRepo.UserDTO{
		ID:          gormUser.ID,
		Username:    gormUser.Username,
		RealName:    gormUser.RealName,
		Email:       gormUser.Email,
		Phone:       gormUser.Phone,
		Status:      gormUser.Status,
		LastLoginAt: gormUser.LastLoginAt,
		CreatedAt:   gormUser.CreatedAt,
		UpdatedAt:   gormUser.UpdatedAt,
	}, nil
}

// FindByUsername 按用户名查询用户（含角色+权限）
func (r *UserReadRepository) FindByUsername(ctx context.Context, username string) (*appRepo.UserWithRolesDTO, error) {
	var gormUser UserGorm
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&gormUser).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrRecordNotFound
		}
		return nil, err
	}

	return r.loadUserWithRoles(ctx, &gormUser)
}

// FindByIDWithRoles 按 ID 查询用户含角色权限
func (r *UserReadRepository) FindByIDWithRoles(ctx context.Context, id string) (*appRepo.UserWithRolesDTO, error) {
	var gormUser UserGorm
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&gormUser).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrRecordNotFound
		}
		return nil, err
	}

	return r.loadUserWithRoles(ctx, &gormUser)
}

// loadUserWithRoles 加载用户角色和权限
func (r *UserReadRepository) loadUserWithRoles(ctx context.Context, gormUser *UserGorm) (*appRepo.UserWithRolesDTO, error) {
	// 查询用户关联的角色
	var userRoles []UserRoleGorm
	if err := r.db.WithContext(ctx).Where("user_id = ?", gormUser.ID).Find(&userRoles).Error; err != nil {
		return nil, err
	}

	// 查询角色详情和权限
	roleEntries := make([]*appRepo.RoleEntry, 0, len(userRoles))
	permSet := make(map[string]bool)
	permEntries := make([]*appRepo.PermissionEntry, 0)

	for _, ur := range userRoles {
		var roleGorm RoleGorm
		if err := r.db.WithContext(ctx).Where("id = ?", ur.RoleID).First(&roleGorm).Error; err != nil {
			continue
		}
		if roleGorm.Status != 1 { // 只包含启用角色
			continue
		}

		roleEntries = append(roleEntries, &appRepo.RoleEntry{
			ID:   roleGorm.ID,
			Name: roleGorm.Name,
			Code: roleGorm.Code,
		})

		// 查询角色关联的权限
		var rolePerms []RolePermissionGorm
		if err := r.db.WithContext(ctx).Where("role_id = ?", ur.RoleID).Find(&rolePerms).Error; err != nil {
			continue
		}

		for _, rp := range rolePerms {
			if permSet[rp.PermissionID] {
				continue
			}
			permSet[rp.PermissionID] = true

			var permGorm PermissionGorm
			if err := r.db.WithContext(ctx).Where("id = ?", rp.PermissionID).First(&permGorm).Error; err != nil {
				continue
			}
			permEntries = append(permEntries, &appRepo.PermissionEntry{
				ID:   permGorm.ID,
				Name: permGorm.Name,
				Code: permGorm.Code,
			})
		}
	}

	if roleEntries == nil {
		roleEntries = make([]*appRepo.RoleEntry, 0)
	}
	if permEntries == nil {
		permEntries = make([]*appRepo.PermissionEntry, 0)
	}

	return &appRepo.UserWithRolesDTO{
		UserDTO: appRepo.UserDTO{
			ID:          gormUser.ID,
			Username:    gormUser.Username,
			RealName:    gormUser.RealName,
			Email:       gormUser.Email,
			Phone:       gormUser.Phone,
			Status:      gormUser.Status,
			LastLoginAt: gormUser.LastLoginAt,
			CreatedAt:   gormUser.CreatedAt,
			UpdatedAt:   gormUser.UpdatedAt,
		},
		Roles:       roleEntries,
		Permissions: permEntries,
	}, nil
}

// List 用户列表（分页+状态筛选）
func (r *UserReadRepository) List(ctx context.Context, offset, limit int, status int) ([]*appRepo.UserDTO, int64, error) {
	query := r.db.WithContext(ctx).Model(&UserGorm{})

	if status >= 0 {
		query = query.Where("status = ?", status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var gormUsers []UserGorm
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&gormUsers).Error; err != nil {
		return nil, 0, err
	}

	dtos := make([]*appRepo.UserDTO, 0, len(gormUsers))
	for _, u := range gormUsers {
		dtos = append(dtos, &appRepo.UserDTO{
			ID:          u.ID,
			Username:    u.Username,
			RealName:    u.RealName,
			Email:       u.Email,
			Phone:       u.Phone,
			Status:      u.Status,
			LastLoginAt: u.LastLoginAt,
			CreatedAt:   u.CreatedAt,
			UpdatedAt:   u.UpdatedAt,
		})
	}

	return dtos, total, nil
}