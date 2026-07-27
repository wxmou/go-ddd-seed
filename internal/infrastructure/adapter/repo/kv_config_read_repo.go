package repo

import (
	"context"

	appRepo "github.com/go-ddd-seed/go-ddd-seed/internal/application/port/repo"
	"github.com/go-ddd-seed/go-ddd-seed/internal/domain"
	"gorm.io/gorm"
)

// KvConfigReadRepository GORM 读仓储实现（CQRS 读模型）
// 直接查询数据库返回 DTO，不经过领域聚合
type KvConfigReadRepository struct {
	db *gorm.DB
}

// NewKvConfigReadRepository 创建读仓储
func NewKvConfigReadRepository(db *gorm.DB) *KvConfigReadRepository {
	return &KvConfigReadRepository{db: db}
}

// toDTO 转换为 DTO
func (g *KvConfigGorm) toDTO() *appRepo.KvConfigDTO {
	return &appRepo.KvConfigDTO{
		ID:          g.ID,
		Key:         g.Key,
		Value:       g.Value,
		Description: g.Description,
		Status:      g.Status,
		CreatedAt:   g.CreatedAt,
		UpdatedAt:   g.UpdatedAt,
	}
}

// FindByID 按 ID 查询
func (r *KvConfigReadRepository) FindByID(ctx context.Context, id string) (*appRepo.KvConfigDTO, error) {
	var gormModel KvConfigGorm
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&gormModel).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrRecordNotFound
		}
		return nil, err
	}
	return gormModel.toDTO(), nil
}

// FindByKey 按 Key 查询
func (r *KvConfigReadRepository) FindByKey(ctx context.Context, key string) (*appRepo.KvConfigDTO, error) {
	var gormModel KvConfigGorm
	err := r.db.WithContext(ctx).Where("key = ?", key).First(&gormModel).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrRecordNotFound
		}
		return nil, err
	}
	return gormModel.toDTO(), nil
}

// List 列表查询（分页+状态筛选, status=-1 表示全部）
func (r *KvConfigReadRepository) List(ctx context.Context, offset, limit int, status int) ([]*appRepo.KvConfigDTO, int64, error) {
	var models []KvConfigGorm
	query := r.db.WithContext(ctx).Model(&KvConfigGorm{})

	if status >= 0 {
		query = query.Where("status = ?", status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, 0, err
	}

	result := make([]*appRepo.KvConfigDTO, 0, len(models))
	for i := range models {
		result = append(result, models[i].toDTO())
	}
	return result, total, nil
}

// Ensure interface compliance
var _ appRepo.KvConfigReadRepository = (*KvConfigReadRepository)(nil)