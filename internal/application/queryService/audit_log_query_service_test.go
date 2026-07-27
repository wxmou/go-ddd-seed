package queryService

import (
	"context"
	"testing"
	"time"

	appRepo "github.com/go-ddd-seed/go-ddd-seed/internal/application/port/repo"
)

// mockAuditLogReadRepo 审计日志读仓储 mock
type mockAuditLogReadRepo struct {
	dtos []*appRepo.AuditLogReadDTO
}

func (m *mockAuditLogReadRepo) List(_ context.Context, query appRepo.AuditLogQuery) ([]*appRepo.AuditLogReadDTO, int64, error) {
	// 模拟过滤
	var filtered []*appRepo.AuditLogReadDTO
	for _, dto := range m.dtos {
		if query.OperatorID != "" && dto.OperatorID != query.OperatorID {
			continue
		}
		if query.Action != "" && dto.Action != query.Action {
			continue
		}
		if query.TargetType != "" && dto.TargetType != query.TargetType {
			continue
		}
		if query.TargetID != "" && dto.TargetID != query.TargetID {
			continue
		}
		if !query.StartTime.IsZero() && dto.CreatedAt.Before(query.StartTime) {
			continue
		}
		if !query.EndTime.IsZero() && dto.CreatedAt.After(query.EndTime) {
			continue
		}
		filtered = append(filtered, dto)
	}

	total := int64(len(filtered))

	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 || query.PageSize > 100 {
		query.PageSize = 20
	}
	offset := (query.Page - 1) * query.PageSize
	if offset >= len(filtered) {
		return []*appRepo.AuditLogReadDTO{}, total, nil
	}
	end := offset + query.PageSize
	if end > len(filtered) {
		end = len(filtered)
	}

	return filtered[offset:end], total, nil
}

func newTestAuditLogQueryService() (*AuditLogQueryService, *mockAuditLogReadRepo) {
	readRepo := &mockAuditLogReadRepo{}
	svc := NewAuditLogQueryService(readRepo)
	return svc, readRepo
}

func TestAuditLogQueryService_List(t *testing.T) {
	now := time.Now()

	t.Run("success", func(t *testing.T) {
		svc, readRepo := newTestAuditLogQueryService()
		readRepo.dtos = []*appRepo.AuditLogReadDTO{
			{ID: "1", OperatorID: "u1", Action: "create", TargetType: "role", CreatedAt: now},
			{ID: "2", OperatorID: "u2", Action: "update", TargetType: "user", CreatedAt: now.Add(-time.Hour)},
		}

		dtos, total, err := svc.List(context.Background(), appRepo.AuditLogQuery{Page: 1, PageSize: 20})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if total != 2 {
			t.Errorf("expected total 2, got %d", total)
		}
		if len(dtos) != 2 {
			t.Errorf("expected 2 items, got %d", len(dtos))
		}
	})

	t.Run("filter by operator_id", func(t *testing.T) {
		svc, readRepo := newTestAuditLogQueryService()
		readRepo.dtos = []*appRepo.AuditLogReadDTO{
			{ID: "1", OperatorID: "u1", Action: "create", TargetType: "role", CreatedAt: now},
			{ID: "2", OperatorID: "u2", Action: "update", TargetType: "user", CreatedAt: now},
		}

		dtos, total, err := svc.List(context.Background(), appRepo.AuditLogQuery{
			OperatorID: "u1",
			Page:       1,
			PageSize:   20,
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if total != 1 {
			t.Errorf("expected total 1, got %d", total)
		}
		if len(dtos) != 1 || dtos[0].ID != "1" {
			t.Errorf("expected item 1, got %+v", dtos)
		}
	})

	t.Run("filter by action", func(t *testing.T) {
		svc, readRepo := newTestAuditLogQueryService()
		readRepo.dtos = []*appRepo.AuditLogReadDTO{
			{ID: "1", OperatorID: "u1", Action: "create", TargetType: "role", CreatedAt: now},
			{ID: "2", OperatorID: "u1", Action: "update", TargetType: "role", CreatedAt: now},
		}

		dtos, total, err := svc.List(context.Background(), appRepo.AuditLogQuery{
			Action:   "create",
			Page:     1,
			PageSize: 20,
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if total != 1 {
			t.Errorf("expected total 1, got %d", total)
		}
		if len(dtos) != 1 || dtos[0].Action != "create" {
			t.Errorf("expected create action, got %+v", dtos[0])
		}
	})

	t.Run("filter by time range", func(t *testing.T) {
		svc, readRepo := newTestAuditLogQueryService()
		readRepo.dtos = []*appRepo.AuditLogReadDTO{
			{ID: "1", OperatorID: "u1", Action: "create", TargetType: "role", CreatedAt: now},
			{ID: "2", OperatorID: "u1", Action: "update", TargetType: "role", CreatedAt: now.Add(-24 * time.Hour)},
		}

		dtos, total, err := svc.List(context.Background(), appRepo.AuditLogQuery{
			StartTime: now.Add(-6 * time.Hour),
			EndTime:   now.Add(time.Hour),
			Page:      1,
			PageSize:  20,
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if total != 1 {
			t.Errorf("expected total 1, got %d", total)
		}
		if len(dtos) != 1 || dtos[0].ID != "1" {
			t.Errorf("expected item 1 (recent), got %+v", dtos[0])
		}
	})

	t.Run("pagination", func(t *testing.T) {
		svc, readRepo := newTestAuditLogQueryService()
		readRepo.dtos = []*appRepo.AuditLogReadDTO{
			{ID: "1", OperatorID: "u1", Action: "create", TargetType: "role", CreatedAt: now},
			{ID: "2", OperatorID: "u1", Action: "update", TargetType: "user", CreatedAt: now},
			{ID: "3", OperatorID: "u2", Action: "delete", TargetType: "role", CreatedAt: now},
		}

		dtos, total, err := svc.List(context.Background(), appRepo.AuditLogQuery{
			Page:     1,
			PageSize: 2,
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if total != 3 {
			t.Errorf("expected total 3, got %d", total)
		}
		if len(dtos) != 2 {
			t.Errorf("expected 2 items on page 1, got %d", len(dtos))
		}
	})

	t.Run("empty result", func(t *testing.T) {
		svc, _ := newTestAuditLogQueryService()

		dtos, total, err := svc.List(context.Background(), appRepo.AuditLogQuery{Page: 1, PageSize: 20})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if total != 0 {
			t.Errorf("expected total 0, got %d", total)
		}
		if len(dtos) != 0 {
			t.Errorf("expected 0 items, got %d", len(dtos))
		}
	})
}