package repo

import (
	"context"
	"time"

	"github.com/go-ddd-seed/go-ddd-seed/internal/domain"
	"gorm.io/gorm"

	domainRepo "github.com/go-ddd-seed/go-ddd-seed/internal/domain/port/repo"
	"github.com/go-ddd-seed/go-ddd-seed/internal/domain/model/kv_config"
)

// Ensure interface compliance
var _ domainRepo.KvConfigRepository = (*KvConfigRepository)(nil)

// KvConfigGorm GORM 模型
type KvConfigGorm struct {
	ID          string    `gorm:"column:id;primaryKey;type:varchar(36)"`
	Key         string    `gorm:"column:key;type:varchar(255);uniqueIndex;not null"`
	Value       string    `gorm:"column:value;type:text;not null"`
	Description string    `gorm:"column:description;type:varchar(500)"`
	Status      int       `gorm:"column:status;default:1"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

// TableName 表名
func (KvConfigGorm) TableName() string {
	return "kv_configs"
}

// fromDomain 从领域模型转换
func fromKvConfigDomain(d *kv_config.KvConfig) *KvConfigGorm {
	return &KvConfigGorm{
		ID:          d.ID,
		Key:         d.Key,
		Value:       d.Value,
		Description: d.Description,
		Status:      d.Status,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
}

// KvConfigRepository GORM 仓储实现（命令侧）
type KvConfigRepository struct {
	RepositoryBase
}

// NewKvConfigRepository 创建仓储
func NewKvConfigRepository(base RepositoryBase) *KvConfigRepository {
	return &KvConfigRepository{RepositoryBase: base}
}

// Save 保存配置（整存整取），自动发布领域事件
func (r *KvConfigRepository) Save(ctx context.Context, config *kv_config.KvConfig) error {
	return r.SaveWithEvents(ctx, config, func(tx *gorm.DB) error {
		gormModel := fromKvConfigDomain(config)
		return tx.Save(gormModel).Error
	})
}

// FindByID 按 ID 查询配置
func (r *KvConfigRepository) FindByID(ctx context.Context, id string) (*kv_config.KvConfig, error) {
	var gormModel KvConfigGorm
	err := r.DB.WithContext(ctx).Where("id = ?", id).First(&gormModel).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrRecordNotFound
		}
		return nil, err
	}

	return &kv_config.KvConfig{
		ID:          gormModel.ID,
		Key:         gormModel.Key,
		Value:       gormModel.Value,
		Description: gormModel.Description,
		Status:      gormModel.Status,
		CreatedAt:   gormModel.CreatedAt,
		UpdatedAt:   gormModel.UpdatedAt,
	}, nil
}

// FindByKey 按 Key 查询配置，用于唯一性校验
func (r *KvConfigRepository) FindByKey(ctx context.Context, key string) (*kv_config.KvConfig, error) {
	var gormModel KvConfigGorm
	err := r.DB.WithContext(ctx).Where("key = ?", key).First(&gormModel).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &kv_config.KvConfig{
		ID:          gormModel.ID,
		Key:         gormModel.Key,
		Value:       gormModel.Value,
		Description: gormModel.Description,
		Status:      gormModel.Status,
		CreatedAt:   gormModel.CreatedAt,
		UpdatedAt:   gormModel.UpdatedAt,
	}, nil
}

// Delete 删除配置
func (r *KvConfigRepository) Delete(ctx context.Context, id string) error {
	result := r.DB.WithContext(ctx).Where("id = ?", id).Delete(&KvConfigGorm{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrRecordNotFound
	}
	return nil
}
