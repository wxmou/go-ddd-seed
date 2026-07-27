package repo

import (
	"context"

	"github.com/go-ddd-seed/go-ddd-seed/internal/domain/model/dict_type"
)

// DictTypeRepository 字典类型仓储接口（命令侧）
// 聚合整存整取，不直接操作条目
type DictTypeRepository interface {
	// FindByID 按 ID 加载完整聚合（含条目列表）
	FindByID(ctx context.Context, id string) (*dict_type.DictType, error)
	// FindByCode 按 Code 加载完整聚合（含条目列表），用于唯一性校验
	FindByCode(ctx context.Context, code string) (*dict_type.DictType, error)
	// FindEntryTypeID 按条目 ID 查找所属类型 ID，用于写操作中定位聚合根
	FindEntryTypeID(ctx context.Context, entryID string) (string, error)
	// Save 保存聚合（类型 + 条目全量替换）
	Save(ctx context.Context, dictType *dict_type.DictType) error
	// Delete 删除聚合（级联删除条目）
	Delete(ctx context.Context, id string) error
}
