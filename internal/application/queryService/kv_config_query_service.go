package queryService

import (
	"context"

	"github.com/go-ddd-seed/go-ddd-seed/internal/application/port/repo"
)

// KvConfigQueryService 键值配置查询服务（CQRS 读模型）
// 直接通过读仓储获取数据，不经过领域聚合的业务规则校验
type KvConfigQueryService struct {
	readRepo repo.KvConfigReadRepository
}

// NewKvConfigQueryService 创建键值配置查询服务
func NewKvConfigQueryService(readRepo repo.KvConfigReadRepository) *KvConfigQueryService {
	return &KvConfigQueryService{readRepo: readRepo}
}

// GetByID 按 ID 查询
func (s *KvConfigQueryService) GetByID(ctx context.Context, id string) (*repo.KvConfigDTO, error) {
	return s.readRepo.FindByID(ctx, id)
}

// GetByKey 按 Key 查询
func (s *KvConfigQueryService) GetByKey(ctx context.Context, key string) (*repo.KvConfigDTO, error) {
	return s.readRepo.FindByKey(ctx, key)
}

// List 列表查询（分页+状态筛选）
func (s *KvConfigQueryService) List(ctx context.Context, page, pageSize, status int) (*PaginatedDTO, error) {
	offset := (page - 1) * pageSize
	list, total, err := s.readRepo.List(ctx, offset, pageSize, status)
	if err != nil {
		return nil, err
	}

	return &PaginatedDTO{
		List:  list,
		Total: total,
		Page:  page,
		Size:  pageSize,
	}, nil
}

// PaginatedDTO 分页 DTO
type PaginatedDTO struct {
	List  any   `json:"list"`
	Total int64 `json:"total"`
	Page  int   `json:"page"`
	Size  int   `json:"size"`
}