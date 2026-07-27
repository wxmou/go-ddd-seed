package queryService

import (
	"context"
	"testing"

	appRepo "github.com/go-ddd-seed/go-ddd-seed/internal/application/port/repo"
	appErrors "github.com/go-ddd-seed/go-ddd-seed/pkg/errors"
)

// ---------- mock ----------

type mockKvConfigReadRepo struct {
	configs map[string]*appRepo.KvConfigDTO
}

func newMockKvConfigReadRepo() *mockKvConfigReadRepo {
	return &mockKvConfigReadRepo{configs: make(map[string]*appRepo.KvConfigDTO)}
}

func (m *mockKvConfigReadRepo) FindByID(_ context.Context, id string) (*appRepo.KvConfigDTO, error) {
	dto, ok := m.configs[id]
	if !ok {
		return nil, appErrors.ErrNotFound
	}
	return dto, nil
}

func (m *mockKvConfigReadRepo) FindByKey(_ context.Context, key string) (*appRepo.KvConfigDTO, error) {
	for _, dto := range m.configs {
		if dto.Key == key {
			return dto, nil
		}
	}
	return nil, appErrors.ErrNotFound
}

func (m *mockKvConfigReadRepo) List(_ context.Context, offset, limit int, status int) ([]*appRepo.KvConfigDTO, int64, error) {
	var filtered []*appRepo.KvConfigDTO
	for _, dto := range m.configs {
		if status < 0 || dto.Status == status {
			filtered = append(filtered, dto)
		}
	}
	total := int64(len(filtered))
	if offset >= len(filtered) {
		return []*appRepo.KvConfigDTO{}, total, nil
	}
	end := offset + limit
	if end > len(filtered) {
		end = len(filtered)
	}
	return filtered[offset:end], total, nil
}

func newTestKvConfigQueryService() (*KvConfigQueryService, *mockKvConfigReadRepo) {
	readRepo := newMockKvConfigReadRepo()
	svc := NewKvConfigQueryService(readRepo)
	return svc, readRepo
}

// ---------- tests ----------

func TestKvConfigQueryService_GetByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, readRepo := newTestKvConfigQueryService()
		readRepo.configs["id-1"] = &appRepo.KvConfigDTO{
			ID: "id-1", Key: "site_name", Value: "TestApp", Status: 1,
		}

		dto, err := svc.GetByID(context.Background(), "id-1")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if dto.Key != "site_name" {
			t.Errorf("expected key 'site_name', got %q", dto.Key)
		}
	})

	t.Run("not found", func(t *testing.T) {
		svc, _ := newTestKvConfigQueryService()
		_, err := svc.GetByID(context.Background(), "nonexistent")
		if err != appErrors.ErrNotFound {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestKvConfigQueryService_GetByKey(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, readRepo := newTestKvConfigQueryService()
		readRepo.configs["id-1"] = &appRepo.KvConfigDTO{
			ID: "id-1", Key: "site_name", Value: "TestApp", Status: 1,
		}

		dto, err := svc.GetByKey(context.Background(), "site_name")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if dto.ID != "id-1" {
			t.Errorf("expected id 'id-1', got %q", dto.ID)
		}
	})

	t.Run("not found", func(t *testing.T) {
		svc, _ := newTestKvConfigQueryService()
		_, err := svc.GetByKey(context.Background(), "nonexistent")
		if err != appErrors.ErrNotFound {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestKvConfigQueryService_List(t *testing.T) {
	svc, readRepo := newTestKvConfigQueryService()
	readRepo.configs["id-1"] = &appRepo.KvConfigDTO{ID: "id-1", Key: "k1", Value: "v1", Status: 1}
	readRepo.configs["id-2"] = &appRepo.KvConfigDTO{ID: "id-2", Key: "k2", Value: "v2", Status: 1}
	readRepo.configs["id-3"] = &appRepo.KvConfigDTO{ID: "id-3", Key: "k3", Value: "v3", Status: 0}

	t.Run("list all", func(t *testing.T) {
		result, err := svc.List(context.Background(), 1, 20, -1)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result.Total != 3 {
			t.Errorf("expected total 3, got %d", result.Total)
		}
	})

	t.Run("filter by status", func(t *testing.T) {
		result, err := svc.List(context.Background(), 1, 20, 1)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result.Total != 2 {
			t.Errorf("expected total 2 (enabled), got %d", result.Total)
		}
	})

	t.Run("pagination", func(t *testing.T) {
		result, err := svc.List(context.Background(), 1, 2, -1)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result.Total != 3 {
			t.Errorf("expected total 3, got %d", result.Total)
		}
		if list, ok := result.List.([]*appRepo.KvConfigDTO); !ok || len(list) != 2 {
			t.Errorf("expected 2 items in list, got %v", result.List)
		}
	})
}