package repo

import (
	"context"
	"time"
)

// DictTypeDTO 字典类型读模型 DTO（CQRS 查询专用）
// 直接映射数据库字段，不包含领域行为
type DictTypeDTO struct {
	ID          string    `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      int       `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// DictEntryDTO 字典条目读模型 DTO
type DictEntryDTO struct {
	ID        string    `json:"id"`
	TypeID    string    `json:"type_id"`
	Label     string    `json:"label"`
	Value     string    `json:"value"`
	SortOrder int       `json:"sort_order"`
	Status    int       `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DictTypeReadRepository 字典类型读仓储接口（CQRS 读模型）
type DictTypeReadRepository interface {
	// FindByID 按 ID 查询类型（不含条目）
	FindByID(ctx context.Context, id string) (*DictTypeDTO, error)
	// FindByCode 按 Code 查询类型
	FindByCode(ctx context.Context, code string) (*DictTypeDTO, error)
	// List 列表查询（分页+状态筛选, status=-1 表示全部）
	List(ctx context.Context, offset, limit int, status int) ([]*DictTypeDTO, int64, error)
	// FindEntriesByTypeID 按类型ID查询条目列表
	FindEntriesByTypeID(ctx context.Context, typeID string) ([]*DictEntryDTO, error)
	// FindEntriesByTypeCode 按类型编码查询已启用条目列表（用于前端获取枚举值）
	FindEntriesByTypeCode(ctx context.Context, code string) ([]*DictEntryDTO, error)
	// FindEntryByID 按条目ID查询单个条目
	FindEntryByID(ctx context.Context, entryID string) (*DictEntryDTO, error)
}