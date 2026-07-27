package repo

import (
	"context"
)

// DictCacheRepository 字典缓存仓储接口
// 抽象 Redis 缓存操作，用于前端枚举值获取
type DictCacheRepository interface {
	// WarmUp 全量缓存预热（启动时调用）
	WarmUp(ctx context.Context) error
	// GetEntriesByCode 按 typeCode 获取已启用条目（优先缓存）
	GetEntriesByCode(ctx context.Context, code string) ([]*DictEntryEntry, error)
	// RefreshByCode 刷新指定 typeCode 的缓存（写操作后调用）
	RefreshByCode(ctx context.Context, code string, entries []*DictEntryEntry) error
	// DeleteByCode 删除指定 typeCode 的缓存（删除类型时调用）
	DeleteByCode(ctx context.Context, code string) error
}

// DictEntryEntry 缓存条目值对象（用于缓存序列化，轻量级）
type DictEntryEntry struct {
	Label string `json:"label"`
	Value string `json:"value"`
}