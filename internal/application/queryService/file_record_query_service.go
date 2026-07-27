package queryService

import (
	"context"

	appRepo "github.com/go-ddd-seed/go-ddd-seed/internal/application/port/repo"
)

// FileRecordQueryService 文件记录查询服务（CQRS 读模型）
// 直接通过读仓储获取数据，不经过领域聚合的业务规则校验
type FileRecordQueryService struct {
	readRepo appRepo.FileRecordReadRepository
}

// NewFileRecordQueryService 创建文件记录查询服务
func NewFileRecordQueryService(readRepo appRepo.FileRecordReadRepository) *FileRecordQueryService {
	return &FileRecordQueryService{readRepo: readRepo}
}

// GetByID 按 ID 查询
func (s *FileRecordQueryService) GetByID(ctx context.Context, id string) (*appRepo.FileRecordReadDTO, error) {
	return s.readRepo.FindByID(ctx, id)
}

// GetByAttach 按业务对象查询关联文件
func (s *FileRecordQueryService) GetByAttach(ctx context.Context, attachType, attachID string) ([]*appRepo.FileRecordReadDTO, error) {
	return s.readRepo.FindByAttach(ctx, attachType, attachID)
}

// List 分页列表
func (s *FileRecordQueryService) List(ctx context.Context, query appRepo.FileRecordQuery) (*PaginatedDTO, error) {
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}

	list, total, err := s.readRepo.List(ctx, query)
	if err != nil {
		return nil, err
	}

	return &PaginatedDTO{
		List:  list,
		Total: total,
		Page:  query.Page,
		Size:  query.PageSize,
	}, nil
}