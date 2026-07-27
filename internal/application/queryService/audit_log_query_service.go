package queryService

import (
	"context"

	appRepo "github.com/go-ddd-seed/go-ddd-seed/internal/application/port/repo"
)

// AuditLogQueryService 审计日志查询服务（读侧）
type AuditLogQueryService struct {
	readRepo appRepo.AuditLogReadRepository
}

// NewAuditLogQueryService 创建审计日志查询服务
func NewAuditLogQueryService(readRepo appRepo.AuditLogReadRepository) *AuditLogQueryService {
	return &AuditLogQueryService{readRepo: readRepo}
}

// List 分页查询审计日志
func (s *AuditLogQueryService) List(ctx context.Context, query appRepo.AuditLogQuery) ([]*appRepo.AuditLogReadDTO, int64, error) {
	return s.readRepo.List(ctx, query)
}