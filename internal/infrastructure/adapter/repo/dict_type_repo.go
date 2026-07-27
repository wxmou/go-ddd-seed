package repo

import (
	"context"
	"time"

	"github.com/go-ddd-seed/go-ddd-seed/internal/domain"
	"gorm.io/gorm"

	domainRepo "github.com/go-ddd-seed/go-ddd-seed/internal/domain/port/repo"
	"github.com/go-ddd-seed/go-ddd-seed/internal/domain/model/dict_type"
)

// Ensure interface compliance
var _ domainRepo.DictTypeRepository = (*DictTypeRepository)(nil)

// DictTypeGorm 字典类型 GORM 模型
type DictTypeGorm struct {
	ID          string    `gorm:"column:id;primaryKey;type:varchar(36)"`
	Code        string    `gorm:"column:code;type:varchar(100);uniqueIndex;not null"`
	Name        string    `gorm:"column:name;type:varchar(200);not null"`
	Description string    `gorm:"column:description;type:varchar(500)"`
	Status      int       `gorm:"column:status;default:1"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

// TableName 表名
func (DictTypeGorm) TableName() string {
	return "dict_types"
}

// DictEntryGorm 字典条目 GORM 模型
type DictEntryGorm struct {
	ID        string    `gorm:"column:id;primaryKey;type:varchar(36)"`
	TypeID    string    `gorm:"column:type_id;type:varchar(36);index;not null"`
	Label     string    `gorm:"column:label;type:varchar(200);not null"`
	Value     string    `gorm:"column:value;type:varchar(200);not null"`
	SortOrder int       `gorm:"column:sort_order;default:0"`
	Status    int       `gorm:"column:status;default:1"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

// TableName 表名
func (DictEntryGorm) TableName() string {
	return "dict_entries"
}

// fromDomainAggregate 从领域聚合转换
func fromDictTypeDomain(d *dict_type.DictType) *DictTypeGorm {
	return &DictTypeGorm{
		ID:          d.ID,
		Code:        d.Code,
		Name:        d.Name,
		Description: d.Description,
		Status:      d.Status,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
}

// fromDomainEntry 从领域实体转换
func fromDictEntryDomain(e *dict_type.DictEntry) *DictEntryGorm {
	return &DictEntryGorm{
		ID:        e.ID,
		TypeID:    "", // 由调用方填充
		Label:     e.Label,
		Value:     e.Value,
		SortOrder: e.SortOrder,
		Status:    e.Status,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}
}

// DictTypeRepository GORM 仓储实现（命令侧）
type DictTypeRepository struct {
	RepositoryBase
}

// NewDictTypeRepository 创建仓储
func NewDictTypeRepository(base RepositoryBase) *DictTypeRepository {
	return &DictTypeRepository{RepositoryBase: base}
}

// FindByID 按 ID 加载完整聚合（含条目列表）
func (r *DictTypeRepository) FindByID(ctx context.Context, id string) (*dict_type.DictType, error) {
	var gormType DictTypeGorm
	if err := r.DB.WithContext(ctx).Where("id = ?", id).First(&gormType).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrRecordNotFound
		}
		return nil, err
	}

	return r.loadDictTypeWithEntries(ctx, &gormType)
}

// FindByCode 按 Code 加载完整聚合（含条目列表），用于唯一性校验
func (r *DictTypeRepository) FindByCode(ctx context.Context, code string) (*dict_type.DictType, error) {
	var gormType DictTypeGorm
	if err := r.DB.WithContext(ctx).Where("code = ?", code).First(&gormType).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return r.loadDictTypeWithEntries(ctx, &gormType)
}

// FindEntryTypeID 按条目 ID 查找所属类型 ID
func (r *DictTypeRepository) FindEntryTypeID(ctx context.Context, entryID string) (string, error) {
	var entry DictEntryGorm
	if err := r.DB.WithContext(ctx).Where("id = ?", entryID).First(&entry).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", domain.ErrRecordNotFound
		}
		return "", err
	}
	return entry.TypeID, nil
}

// loadDictTypeWithEntries 加载字典类型及其条目列表
func (r *DictTypeRepository) loadDictTypeWithEntries(ctx context.Context, gormType *DictTypeGorm) (*dict_type.DictType, error) {
	// 加载条目列表
	var gormEntries []DictEntryGorm
	if err := r.DB.WithContext(ctx).
		Where("type_id = ?", gormType.ID).
		Order("sort_order ASC, created_at ASC").
		Find(&gormEntries).Error; err != nil {
		return nil, err
	}

	entries := make([]*dict_type.DictEntry, 0, len(gormEntries))
	for _, ge := range gormEntries {
		entries = append(entries, &dict_type.DictEntry{
			ID:        ge.ID,
			Label:     ge.Label,
			Value:     ge.Value,
			SortOrder: ge.SortOrder,
			Status:    ge.Status,
			CreatedAt: ge.CreatedAt,
			UpdatedAt: ge.UpdatedAt,
		})
	}

	return &dict_type.DictType{
		ID:          gormType.ID,
		Code:        gormType.Code,
		Name:        gormType.Name,
		Description: gormType.Description,
		Status:      gormType.Status,
		Entries:     entries,
		CreatedAt:   gormType.CreatedAt,
		UpdatedAt:   gormType.UpdatedAt,
	}, nil
}

// Save 保存字典类型（整存整取：类型 + 条目全量替换），自动发布领域事件
func (r *DictTypeRepository) Save(ctx context.Context, dictType *dict_type.DictType) error {
	return r.SaveWithEvents(ctx, dictType, func(tx *gorm.DB) error {
		// 保存类型基本信息
		gormType := fromDictTypeDomain(dictType)
		if err := tx.Save(gormType).Error; err != nil {
			return err
		}

		// 全量替换条目：删除旧条目，插入新条目
		if err := tx.Where("type_id = ?", dictType.ID).Delete(&DictEntryGorm{}).Error; err != nil {
			return err
		}

		for _, entry := range dictType.Entries {
			gormEntry := fromDictEntryDomain(entry)
			gormEntry.TypeID = dictType.ID
			if err := tx.Create(gormEntry).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// Delete 删除字典类型（级联删除条目）
func (r *DictTypeRepository) Delete(ctx context.Context, id string) error {
	return r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 先删除条目
		if err := tx.Where("type_id = ?", id).Delete(&DictEntryGorm{}).Error; err != nil {
			return err
		}

		// 再删除类型
		result := tx.Where("id = ?", id).Delete(&DictTypeGorm{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return domain.ErrRecordNotFound
		}
		return nil
	})
}
