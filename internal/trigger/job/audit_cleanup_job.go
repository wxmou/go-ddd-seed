package job

import (
	"context"
	"time"

	appApi "github.com/go-ddd-seed/go-ddd-seed/internal/application/port/thirdPartyApi"
)

// Ensure interface compliance
var _ appApi.Job = (*AuditCleanupJob)(nil)

// AuditCleanupJob 审计日志清理任务
// 删除超过指定保留天数的审计日志
type AuditCleanupJob struct {
	repo          appApi.AuditLogRepository
	retentionDays int // 保留天数，默认 180
}

// NewAuditCleanupJob 创建审计日志清理任务
func NewAuditCleanupJob(repo appApi.AuditLogRepository, retentionDays int) *AuditCleanupJob {
	if retentionDays <= 0 {
		retentionDays = 180
	}
	return &AuditCleanupJob{repo: repo, retentionDays: retentionDays}
}

// Name 返回任务名称
func (j *AuditCleanupJob) Name() string {
	return "audit_cleanup"
}

// Run 执行清理
func (j *AuditCleanupJob) Run(ctx context.Context) error {
	cutoff := time.Now().AddDate(0, 0, -j.retentionDays)
	_, err := j.repo.DeleteOlderThan(ctx, cutoff)
	return err
}