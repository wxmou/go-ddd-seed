package repo

import (
	"context"

	appRepo "github.com/go-ddd-seed/go-ddd-seed/internal/application/port/repo"
	"github.com/go-ddd-seed/go-ddd-seed/internal/domain"
	"gorm.io/gorm"
)

// Ensure interface compliance
var _ appRepo.DictTypeReadRepository = (*DictTypeReadRepository)(nil)

// DictTypeReadRepository GORM 读仓储实现（CQRS 读模型）
type DictTypeReadRepository struct {
	db *gorm.DB
}

// NewDictTypeReadRepository 创建读仓储
func NewDictTypeReadRepository(db *gorm.DB) *DictTypeReadRepository {
	return &DictTypeReadRepository{db: db}
}

// toTypeDTO 转换为类型 DTO
func (g *DictTypeGorm) toTypeDTO() *appRepo.DictTypeDTO {
	return &appRepo.DictTypeDTO{
		ID:          g.ID,
		Code:        g.Code,
		Name:        g.Name,
		Description: g.Description,
		Status:      g.Status,
		CreatedAt:   g.CreatedAt,
		UpdatedAt:   g.UpdatedAt,
	}
}

// toEntryDTO 转换为条目 DTO
func (g *DictEntryGorm) toEntryDTO() *appRepo.DictEntryDTO {
	return &appRepo.DictEntryDTO{
		ID:        g.ID,
		TypeID:    g.TypeID,
		Label:     g.Label,
		Value:     g.Value,
		SortOrder: g.SortOrder,
		Status:    g.Status,
		CreatedAt: g.CreatedAt,
		UpdatedAt: g.UpdatedAt,
	}
}

// FindByID 按 ID 查询类型
func (r *DictTypeReadRepository) FindByID(ctx context.Context, id string) (*appRepo.DictTypeDTO, error) {
	var gormModel DictTypeGorm
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&gormModel).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrRecordNotFound
		}
		return nil, err
	}
	return gormModel.toTypeDTO(), nil
}

// FindByCode 按 Code 查询类型
func (r *DictTypeReadRepository) FindByCode(ctx context.Context, code string) (*appRepo.DictTypeDTO, error) {
	var gormModel DictTypeGorm
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&gormModel).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrRecordNotFound
		}
		return nil, err
	}
	return gormModel.toTypeDTO(), nil
}

// List 列表查询（分页+状态筛选, status=-1 表示全部）
func (r *DictTypeReadRepository) List(ctx context.Context, offset, limit int, status int) ([]*appRepo.DictTypeDTO, int64, error) {
	var models []DictTypeGorm
	query := r.db.WithContext(ctx).Model(&DictTypeGorm{})

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

	result := make([]*appRepo.DictTypeDTO, 0, len(models))
	for i := range models {
		result = append(result, models[i].toTypeDTO())
	}
	return result, total, nil
}

// FindEntriesByTypeID 按类型 ID 查询条目列表
func (r *DictTypeReadRepository) FindEntriesByTypeID(ctx context.Context, typeID string) ([]*appRepo.DictEntryDTO, error) {
	var models []DictEntryGorm
	err := r.db.WithContext(ctx).
		Where("type_id = ?", typeID).
		Order("sort_order ASC, created_at ASC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	result := make([]*appRepo.DictEntryDTO, 0, len(models))
	for i := range models {
		result = append(result, models[i].toEntryDTO())
	}
	return result, nil
}

// FindEntriesByTypeCode 按类型编码查询已启用条目列表
func (r *DictTypeReadRepository) FindEntriesByTypeCode(ctx context.Context, code string) ([]*appRepo.DictEntryDTO, error) {
	var models []DictEntryGorm
	err := r.db.WithContext(ctx).
		Joins("JOIN dict_types ON dict_entries.type_id = dict_types.id").
		Where("dict_types.code = ?", code).
		Where("dict_entries.status = ?", 1).
		Where("dict_types.status = ?", 1).
		Order("dict_entries.sort_order ASC, dict_entries.created_at ASC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	result := make([]*appRepo.DictEntryDTO, 0, len(models))
	for i := range models {
		result = append(result, models[i].toEntryDTO())
	}
	return result, nil
}

// FindEntryByID 按条目 ID 查询单个条目
func (r *DictTypeReadRepository) FindEntryByID(ctx context.Context, entryID string) (*appRepo.DictEntryDTO, error) {
	var gormModel DictEntryGorm
	err := r.db.WithContext(ctx).Where("id = ?", entryID).First(&gormModel).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrRecordNotFound
		}
		return nil, err
	}
	return gormModel.toEntryDTO(), nil
}
