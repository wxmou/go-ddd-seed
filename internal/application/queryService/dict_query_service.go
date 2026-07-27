package queryService

import (
	"context"

	"github.com/go-ddd-seed/go-ddd-seed/internal/application/port/repo"
)

// DictQueryService 字典查询服务（CQRS 读模型）
// 类型查询走读仓储，前端枚举值获取走缓存仓储
type DictQueryService struct {
	readRepo  repo.DictTypeReadRepository
	cacheRepo repo.DictCacheRepository
}

// NewDictQueryService 创建字典查询服务
func NewDictQueryService(readRepo repo.DictTypeReadRepository, cacheRepo repo.DictCacheRepository) *DictQueryService {
	return &DictQueryService{readRepo: readRepo, cacheRepo: cacheRepo}
}

// GetByID 按 ID 查询类型
func (s *DictQueryService) GetByID(ctx context.Context, id string) (*repo.DictTypeDTO, error) {
	return s.readRepo.FindByID(ctx, id)
}

// GetByCode 按 Code 查询类型
func (s *DictQueryService) GetByCode(ctx context.Context, code string) (*repo.DictTypeDTO, error) {
	return s.readRepo.FindByCode(ctx, code)
}

// List 类型列表查询（分页+状态筛选）
func (s *DictQueryService) List(ctx context.Context, page, pageSize, status int) (*PaginatedDTO, error) {
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

// GetEntriesByTypeID 按类型 ID 查询条目列表
func (s *DictQueryService) GetEntriesByTypeID(ctx context.Context, typeID string) ([]*repo.DictEntryDTO, error) {
	return s.readRepo.FindEntriesByTypeID(ctx, typeID)
}

// GetEntriesByCode 按 typeCode 获取已启用条目（委托缓存仓储处理 Cache-Aside）
func (s *DictQueryService) GetEntriesByCode(ctx context.Context, code string) ([]*repo.DictEntryEntry, error) {
	return s.cacheRepo.GetEntriesByCode(ctx, code)
}

// WarmUpCache 缓存预热
func (s *DictQueryService) WarmUpCache(ctx context.Context) error {
	return s.cacheRepo.WarmUp(ctx)
}

// GetEntryByID 按条目 ID 查询
func (s *DictQueryService) GetEntryByID(ctx context.Context, entryID string) (*repo.DictEntryDTO, error) {
	return s.readRepo.FindEntryByID(ctx, entryID)
}
