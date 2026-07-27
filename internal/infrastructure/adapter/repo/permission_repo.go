package repo

import (
	"context"
	"time"

	"github.com/go-ddd-seed/go-ddd-seed/internal/domain"
	"gorm.io/gorm"

	domainRepo "github.com/go-ddd-seed/go-ddd-seed/internal/domain/port/repo"
	"github.com/go-ddd-seed/go-ddd-seed/internal/domain/model/permission"
)

// Ensure interface compliance
var _ domainRepo.PermissionRepository = (*PermissionRepository)(nil)

// PermissionGorm 权限 GORM 模型
type PermissionGorm struct {
	ID          string    `gorm:"column:id;primaryKey;type:varchar(36)"`
	Name        string    `gorm:"column:name;type:varchar(200);not null"`
	Code        string    `gorm:"column:code;type:varchar(200);uniqueIndex;not null"`
	Description string    `gorm:"column:description;type:varchar(500)"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

// TableName 表名
func (PermissionGorm) TableName() string {
	return "permissions"
}

func fromPermissionDomain(p *permission.Permission) *PermissionGorm {
	return &PermissionGorm{
		ID:          p.ID,
		Name:        p.Name,
		Code:        p.Code,
		Description: p.Description,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

// PermissionRepository GORM 权限仓储实现（命令侧）
type PermissionRepository struct {
	RepositoryBase
}

// NewPermissionRepository 创建权限仓储
func NewPermissionRepository(base RepositoryBase) *PermissionRepository {
	return &PermissionRepository{RepositoryBase: base}
}

// Save 保存权限，自动发布领域事件
func (r *PermissionRepository) Save(ctx context.Context, permission *permission.Permission) error {
	return r.SaveWithEvents(ctx, permission, func(tx *gorm.DB) error {
		gormPerm := fromPermissionDomain(permission)
		return tx.Save(gormPerm).Error
	})
}

// FindByID 按 ID 查询
func (r *PermissionRepository) FindByID(ctx context.Context, id string) (*permission.Permission, error) {
	var gormPerm PermissionGorm
	if err := r.DB.WithContext(ctx).Where("id = ?", id).First(&gormPerm).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrRecordNotFound
		}
		return nil, err
	}

	return &permission.Permission{
		ID:          gormPerm.ID,
		Name:        gormPerm.Name,
		Code:        gormPerm.Code,
		Description: gormPerm.Description,
		CreatedAt:   gormPerm.CreatedAt,
		UpdatedAt:   gormPerm.UpdatedAt,
	}, nil
}

// FindByCode 按 Code 查询
func (r *PermissionRepository) FindByCode(ctx context.Context, code string) (*permission.Permission, error) {
	var gormPerm PermissionGorm
	if err := r.DB.WithContext(ctx).Where("code = ?", code).First(&gormPerm).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &permission.Permission{
		ID:          gormPerm.ID,
		Name:        gormPerm.Name,
		Code:        gormPerm.Code,
		Description: gormPerm.Description,
		CreatedAt:   gormPerm.CreatedAt,
		UpdatedAt:   gormPerm.UpdatedAt,
	}, nil
}

// Delete 删除权限
func (r *PermissionRepository) Delete(ctx context.Context, id string) error {
	result := r.DB.WithContext(ctx).Where("id = ?", id).Delete(&PermissionGorm{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrRecordNotFound
	}
	return nil
}