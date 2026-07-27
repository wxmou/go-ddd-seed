package repo

import (
	"context"

	"github.com/go-ddd-seed/go-ddd-seed/internal/domain/model/kv_config"
)

// KvConfigRepository 键值配置仓储接口（命令侧）
// 只包含写操作和必要的领域对象加载，纯查询操作请使用 KvConfigReadRepository
type KvConfigRepository interface {
	Save(ctx context.Context, config *kv_config.KvConfig) error
	FindByID(ctx context.Context, id string) (*kv_config.KvConfig, error)
	// FindByKey 按 Key 加载配置，用于唯一性校验
	FindByKey(ctx context.Context, key string) (*kv_config.KvConfig, error)
	Delete(ctx context.Context, id string) error
}