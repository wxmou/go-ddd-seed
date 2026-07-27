package repo

import (
	"context"

	"github.com/go-ddd-seed/go-ddd-seed/internal/domain"
	"gorm.io/gorm"

	appRepo "github.com/go-ddd-seed/go-ddd-seed/internal/application/port/repo"
)

// Ensure interface compliance
var _ appRepo.RoleReadRepository = (*RoleReadRepository)(nil)

// RoleReadRepository 角色读仓储 GORM 实现（CQRS 读模型）
// 与 RoleRepository（命令侧）分开，职责单一
type RoleReadRepository struct {
	db *gorm.DB
}

// NewRoleReadRepository 创建角色读仓储
func NewRoleReadRepository(db *gorm.DB) *RoleReadRepository {
	return &RoleReadRepository{db: db}
}

// FindByID 按 ID 查询角色
func (r *RoleReadRepository) FindByID(ctx context.Context, id string) (*appRepo.RoleDTO, error) {
	var gormRole RoleGorm
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&gormRole).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrRecordNotFound
		}
		return nil, err
	}

	return &appRepo.RoleDTO{
		ID:          gormRole.ID,
		Name:        gormRole.Name,
		Code:        gormRole.Code,
		Description: gormRole.Description,
		Status:      gormRole.Status,
		CreatedAt:   gormRole.CreatedAt,
		UpdatedAt:   gormRole.UpdatedAt,
	}, nil
}

// FindByCode 按 Code 查询角色（含权限）
func (r *RoleReadRepository) FindByCode(ctx context.Context, code string) (*appRepo.RoleWithPermissionsDTO, error) {
	var gormRole RoleGorm
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&gormRole).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrRecordNotFound
		}
		return nil, err
	}

	return r.loadRoleWithPermDTOs(ctx, &gormRole)
}

// List 角色列表
func (r *RoleReadRepository) List(ctx context.Context) ([]*appRepo.RoleDTO, error) {
	var gormRoles []RoleGorm
	if err := r.db.WithContext(ctx).Order("created_at ASC").Find(&gormRoles).Error; err != nil {
		return nil, err
	}

	dtos := make([]*appRepo.RoleDTO, 0, len(gormRoles))
	for _, gr := range gormRoles {
		dtos = append(dtos, &appRepo.RoleDTO{
			ID:          gr.ID,
			Name:        gr.Name,
			Code:        gr.Code,
			Description: gr.Description,
			Status:      gr.Status,
			CreatedAt:   gr.CreatedAt,
			UpdatedAt:   gr.UpdatedAt,
		})
	}

	return dtos, nil
}

func (r *RoleReadRepository) loadRoleWithPermDTOs(ctx context.Context, gormRole *RoleGorm) (*appRepo.RoleWithPermissionsDTO, error) {
	var gormPerms []RolePermissionGorm
	if err := r.db.WithContext(ctx).Where("role_id = ?", gormRole.ID).Find(&gormPerms).Error; err != nil {
		return nil, err
	}

	permDTOs := make([]*appRepo.PermissionDTO, 0, len(gormPerms))
	for _, gp := range gormPerms {
		var permGorm PermissionGorm
		if err := r.db.WithContext(ctx).Where("id = ?", gp.PermissionID).First(&permGorm).Error; err != nil {
			continue
		}
		permDTOs = append(permDTOs, &appRepo.PermissionDTO{
			ID:          permGorm.ID,
			Name:        permGorm.Name,
			Code:        permGorm.Code,
			Description: permGorm.Description,
			CreatedAt:   permGorm.CreatedAt,
			UpdatedAt:   permGorm.UpdatedAt,
		})
	}

	if permDTOs == nil {
		permDTOs = make([]*appRepo.PermissionDTO, 0)
	}

	return &appRepo.RoleWithPermissionsDTO{
		RoleDTO: appRepo.RoleDTO{
			ID:          gormRole.ID,
			Name:        gormRole.Name,
			Code:        gormRole.Code,
			Description: gormRole.Description,
			Status:      gormRole.Status,
			CreatedAt:   gormRole.CreatedAt,
			UpdatedAt:   gormRole.UpdatedAt,
		},
		Permissions: permDTOs,
	}, nil
}