package repo

import (
	"context"
	"time"
)

// KvConfigDTO 键值配置读模型 DTO（CQRS 查询专用）
// 直接映射数据库字段，不包含领域行为
type KvConfigDTO struct {
	ID          string    `json:"id"`
	Key         string    `json:"key"`
	Value       string    `json:"value"`
	Description string    `json:"description"`
	Status      int       `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// KvConfigReadRepository 键值配置读仓储接口（CQRS 读模型）
// 专用于查询操作，不参与领域聚合的写流程
type KvConfigReadRepository interface {
	// FindByID 按 ID 查询
	FindByID(ctx context.Context, id string) (*KvConfigDTO, error)
	// FindByKey 按 Key 查询
	FindByKey(ctx context.Context, key string) (*KvConfigDTO, error)
	// List 列表查询（分页+状态筛选, status=-1 表示全部）
	List(ctx context.Context, offset, limit int, status int) ([]*KvConfigDTO, int64, error)
}