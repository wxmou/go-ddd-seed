package repo

import (
	"context"

	"github.com/go-ddd-seed/go-ddd-seed/internal/domain"
	"gorm.io/gorm"

	appRepo "github.com/go-ddd-seed/go-ddd-seed/internal/application/port/repo"
)

// Ensure interface compliance
var _ appRepo.PermissionReadRepository = (*PermissionReadRepository)(nil)

// PermissionReadRepository 权限读仓储 GORM 实现（CQRS 读模型）
// 与 PermissionRepository（命令侧）分开，职责单一
type PermissionReadRepository struct {
	db *gorm.DB
}

// NewPermissionReadRepository 创建权限读仓储
func NewPermissionReadRepository(db *gorm.DB) *PermissionReadRepository {
	return &PermissionReadRepository{db: db}
}

// FindByCode 按 Code 查询权限
func (r *PermissionReadRepository) FindByCode(ctx context.Context, code string) (*appRepo.PermissionDTO, error) {
	var gormPerm PermissionGorm
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&gormPerm).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrRecordNotFound
		}
		return nil, err
	}

	return &appRepo.PermissionDTO{
		ID:          gormPerm.ID,
		Name:        gormPerm.Name,
		Code:        gormPerm.Code,
		Description: gormPerm.Description,
		CreatedAt:   gormPerm.CreatedAt,
		UpdatedAt:   gormPerm.UpdatedAt,
	}, nil
}

// List 权限列表
func (r *PermissionReadRepository) List(ctx context.Context) ([]*appRepo.PermissionDTO, error) {
	var gormPerms []PermissionGorm
	if err := r.db.WithContext(ctx).Order("created_at ASC").Find(&gormPerms).Error; err != nil {
		return nil, err
	}

	dtos := make([]*appRepo.PermissionDTO, 0, len(gormPerms))
	for _, p := range gormPerms {
		dtos = append(dtos, &appRepo.PermissionDTO{
			ID:          p.ID,
			Name:        p.Name,
			Code:        p.Code,
			Description: p.Description,
			CreatedAt:   p.CreatedAt,
			UpdatedAt:   p.UpdatedAt,
		})
	}

	return dtos, nil
}