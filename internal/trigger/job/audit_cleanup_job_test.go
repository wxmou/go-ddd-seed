package job

import (
	"context"
	"testing"
	"time"

	appApi "github.com/go-ddd-seed/go-ddd-seed/internal/application/port/thirdPartyApi"
)

// mockAuditCleanupRepo 审计日志清理 mock 仓储
type mockAuditCleanupRepo struct {
	deletedCount int64
}

func (m *mockAuditCleanupRepo) Save(_ context.Context, _ *appApi.AuditLogDTO) error {
	return nil
}

func (m *mockAuditCleanupRepo) DeleteOlderThan(_ context.Context, before time.Time) (int64, error) {
	m.deletedCount = 10
	return m.deletedCount, nil
}

func TestAuditCleanupJob(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := &mockAuditCleanupRepo{}
		job := NewAuditCleanupJob(repo, 180)

		err := job.Run(context.Background())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("default retention days", func(t *testing.T) {
		repo := &mockAuditCleanupRepo{}
		job := NewAuditCleanupJob(repo, 0) // 0 means default

		err := job.Run(context.Background())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("custom retention days", func(t *testing.T) {
		repo := &mockAuditCleanupRepo{}
		job := NewAuditCleanupJob(repo, 90)

		err := job.Run(context.Background())
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
}

func TestAuditCleanupJob_ImplementsJob(t *testing.T) {
	repo := &mockAuditCleanupRepo{}
	job := NewAuditCleanupJob(repo, 180)
	if job.Name() != "audit_cleanup" {
		t.Errorf("expected name 'audit_cleanup', got %q", job.Name())
	}

	// compile-time check
	var _ appApi.Job = job
}