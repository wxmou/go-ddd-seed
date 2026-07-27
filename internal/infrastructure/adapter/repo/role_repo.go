package repo

import (
	"context"
	"time"

	"github.com/go-ddd-seed/go-ddd-seed/internal/domain"
	"gorm.io/gorm"

	domainRepo "github.com/go-ddd-seed/go-ddd-seed/internal/domain/port/repo"
	"github.com/go-ddd-seed/go-ddd-seed/internal/domain/model/role"
	
)

// Ensure interface compliance
var _ domainRepo.RoleRepository = (*RoleRepository)(nil)

// RoleGorm 角色 GORM 模型
type RoleGorm struct {
	ID          string    `gorm:"column:id;primaryKey;type:varchar(36)"`
	Name        string    `gorm:"column:name;type:varchar(100);not null"`
	Code        string    `gorm:"column:code;type:varchar(100);uniqueIndex;not null"`
	Description string    `gorm:"column:description;type:varchar(500)"`
	Status      int       `gorm:"column:status;default:1"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

// TableName 表名
func (RoleGorm) TableName() string {
	return "roles"
}

// RolePermissionGorm 角色-权限关联 GORM 模型
type RolePermissionGorm struct {
	ID           string    `gorm:"column:id;primaryKey;type:varchar(36)"`
	RoleID       string    `gorm:"column:role_id;type:varchar(36);uniqueIndex:uk_role_permission;not null"`
	PermissionID string    `gorm:"column:permission_id;type:varchar(36);uniqueIndex:uk_role_permission;not null"`
	CreatedAt    time.Time `gorm:"column:created_at"`
}

// TableName 表名
func (RolePermissionGorm) TableName() string {
	return "role_permissions"
}

func fromRoleDomain(r *role.Role) *RoleGorm {
	return &RoleGorm{
		ID:          r.ID,
		Name:        r.Name,
		Code:        r.Code,
		Description: r.Description,
		Status:      r.Status,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

func fromRolePermissionDomain(rp *role.RolePermission) *RolePermissionGorm {
	return &RolePermissionGorm{
		ID:           rp.ID,
		RoleID:       rp.RoleID,
		PermissionID: rp.PermissionID,
		CreatedAt:    rp.CreatedAt,
	}
}

// RoleRepository GORM 仓储实现（命令侧）
type RoleRepository struct {
	RepositoryBase
}

// NewRoleRepository 创建角色仓储
func NewRoleRepository(base RepositoryBase) *RoleRepository {
	return &RoleRepository{RepositoryBase: base}
}

// Save 保存角色（含权限关联），自动发布领域事件
func (r *RoleRepository) Save(ctx context.Context, role *role.Role) error {
	return r.SaveWithEvents(ctx, role, func(tx *gorm.DB) error {
		gormRole := fromRoleDomain(role)
		if err := tx.Save(gormRole).Error; err != nil {
			return err
		}

		// 全量替换权限关联
		if err := tx.Where("role_id = ?", role.ID).Delete(&RolePermissionGorm{}).Error; err != nil {
			return err
		}

		for _, rp := range role.Permissions {
			gormRP := fromRolePermissionDomain(rp)
			gormRP.RoleID = role.ID
			if err := tx.Create(gormRP).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// FindByID 按 ID 加载角色（含权限关联）
func (r *RoleRepository) FindByID(ctx context.Context, id string) (*role.Role, error) {
	var gormRole RoleGorm
	if err := r.DB.WithContext(ctx).Where("id = ?", id).First(&gormRole).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrRecordNotFound
		}
		return nil, err
	}

	return r.loadRoleWithPermissions(ctx, &gormRole)
}

// FindByCode 按 Code 加载角色（含权限关联）
func (r *RoleRepository) FindByCode(ctx context.Context, code string) (*role.Role, error) {
	var gormRole RoleGorm
	if err := r.DB.WithContext(ctx).Where("code = ?", code).First(&gormRole).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrRecordNotFound
		}
		return nil, err
	}

	return r.loadRoleWithPermissions(ctx, &gormRole)
}

// Delete 删除角色
func (r *RoleRepository) Delete(ctx context.Context, id string) error {
	return r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", id).Delete(&RolePermissionGorm{}).Error; err != nil {
			return err
		}
		result := tx.Where("id = ?", id).Delete(&RoleGorm{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return domain.ErrRecordNotFound
		}
		return nil
	})
}

// loadRoleWithPermissions 加载角色的权限关联列表
func (r *RoleRepository) loadRoleWithPermissions(ctx context.Context, gormRole *RoleGorm) (*role.Role, error) {
	var gormPerms []RolePermissionGorm
	if err := r.DB.WithContext(ctx).Where("role_id = ?", gormRole.ID).Find(&gormPerms).Error; err != nil {
		return nil, err
	}

	perms := make([]*role.RolePermission, 0, len(gormPerms))
	for _, gp := range gormPerms {
		perms = append(perms, &role.RolePermission{
			ID:           gp.ID,
			RoleID:       gp.RoleID,
			PermissionID: gp.PermissionID,
			CreatedAt:    gp.CreatedAt,
		})
	}

	return &role.Role{
		ID:          gormRole.ID,
		Name:        gormRole.Name,
		Code:        gormRole.Code,
		Description: gormRole.Description,
		Status:      gormRole.Status,
		Permissions: perms,
		CreatedAt:   gormRole.CreatedAt,
		UpdatedAt:   gormRole.UpdatedAt,
	}, nil
}