package queryService

import (
	"context"
	"testing"

	appRepo "github.com/go-ddd-seed/go-ddd-seed/internal/application/port/repo"
	appErrors "github.com/go-ddd-seed/go-ddd-seed/pkg/errors"
)

// ---------- mocks (query-specific) ----------

type queryMockReadRepo struct {
	types   map[string]*appRepo.DictTypeDTO
	entries map[string]*appRepo.DictEntryDTO
}

func newQueryMockReadRepo() *queryMockReadRepo {
	return &queryMockReadRepo{
		types:   make(map[string]*appRepo.DictTypeDTO),
		entries: make(map[string]*appRepo.DictEntryDTO),
	}
}

func (m *queryMockReadRepo) FindByID(_ context.Context, id string) (*appRepo.DictTypeDTO, error) {
	dto, ok := m.types[id]
	if !ok {
		return nil, appErrors.ErrNotFound
	}
	return dto, nil
}

func (m *queryMockReadRepo) FindByCode(_ context.Context, code string) (*appRepo.DictTypeDTO, error) {
	for _, dto := range m.types {
		if dto.Code == code {
			return dto, nil
		}
	}
	return nil, appErrors.ErrNotFound
}

func (m *queryMockReadRepo) List(_ context.Context, offset, limit int, status int) ([]*appRepo.DictTypeDTO, int64, error) {
	var filtered []*appRepo.DictTypeDTO
	for _, dto := range m.types {
		if status < 0 || dto.Status == status {
			filtered = append(filtered, dto)
		}
	}
	total := int64(len(filtered))
	if offset >= len(filtered) {
		return []*appRepo.DictTypeDTO{}, total, nil
	}
	end := offset + limit
	if end > len(filtered) {
		end = len(filtered)
	}
	return filtered[offset:end], total, nil
}

func (m *queryMockReadRepo) FindEntriesByTypeID(_ context.Context, typeID string) ([]*appRepo.DictEntryDTO, error) {
	var result []*appRepo.DictEntryDTO
	for _, e := range m.entries {
		if e.TypeID == typeID {
			result = append(result, e)
		}
	}
	return result, nil
}

func (m *queryMockReadRepo) FindEntriesByTypeCode(_ context.Context, code string) ([]*appRepo.DictEntryDTO, error) {
	var typeDTO *appRepo.DictTypeDTO
	for _, dto := range m.types {
		if dto.Code == code {
			typeDTO = dto
			break
		}
	}
	if typeDTO == nil {
		return nil, appErrors.ErrNotFound
	}
	var result []*appRepo.DictEntryDTO
	for _, e := range m.entries {
		if e.TypeID == typeDTO.ID && e.Status == 1 {
			result = append(result, e)
		}
	}
	return result, nil
}

func (m *queryMockReadRepo) FindEntryByID(_ context.Context, entryID string) (*appRepo.DictEntryDTO, error) {
	e, ok := m.entries[entryID]
	if !ok {
		return nil, appErrors.ErrNotFound
	}
	return e, nil
}

type queryMockCacheRepo struct {
	cache map[string][]*appRepo.DictEntryEntry
}

func newQueryMockCacheRepo() *queryMockCacheRepo {
	return &queryMockCacheRepo{cache: make(map[string][]*appRepo.DictEntryEntry)}
}

func (m *queryMockCacheRepo) WarmUp(_ context.Context) error {
	return nil
}

func (m *queryMockCacheRepo) GetEntriesByCode(_ context.Context, code string) ([]*appRepo.DictEntryEntry, error) {
	entries, ok := m.cache[code]
	if !ok {
		return nil, nil // cache miss
	}
	return entries, nil
}

func (m *queryMockCacheRepo) RefreshByCode(_ context.Context, code string, entries []*appRepo.DictEntryEntry) error {
	if len(entries) == 0 {
		delete(m.cache, code)
		return nil
	}
	m.cache[code] = entries
	return nil
}

func (m *queryMockCacheRepo) DeleteByCode(_ context.Context, code string) error {
	delete(m.cache, code)
	return nil
}

func newQueryTestService() (*DictQueryService, *queryMockReadRepo, *queryMockCacheRepo) {
	readRepo := newQueryMockReadRepo()
	cacheRepo := newQueryMockCacheRepo()
	svc := NewDictQueryService(readRepo, cacheRepo)
	return svc, readRepo, cacheRepo
}

// ---------- tests ----------

func TestDictQueryService_GetByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, readRepo, _ := newQueryTestService()
		readRepo.types["id-1"] = &appRepo.DictTypeDTO{
			ID: "id-1", Code: "gender", Name: "性别", Status: 1,
		}

		dto, err := svc.GetByID(context.Background(), "id-1")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if dto.Code != "gender" {
			t.Errorf("expected code 'gender', got %q", dto.Code)
		}
	})

	t.Run("not found", func(t *testing.T) {
		svc, _, _ := newQueryTestService()
		_, err := svc.GetByID(context.Background(), "nonexistent")
		if err != appErrors.ErrNotFound {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestDictQueryService_GetByCode(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, readRepo, _ := newQueryTestService()
		readRepo.types["id-1"] = &appRepo.DictTypeDTO{
			ID: "id-1", Code: "gender", Name: "性别", Status: 1,
		}

		dto, err := svc.GetByCode(context.Background(), "gender")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if dto.ID != "id-1" {
			t.Errorf("expected id 'id-1', got %q", dto.ID)
		}
	})

	t.Run("not found", func(t *testing.T) {
		svc, _, _ := newQueryTestService()
		_, err := svc.GetByCode(context.Background(), "nonexistent")
		if err != appErrors.ErrNotFound {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestDictQueryService_List(t *testing.T) {
	svc, readRepo, _ := newQueryTestService()
	readRepo.types["id-1"] = &appRepo.DictTypeDTO{ID: "id-1", Code: "gender", Name: "性别", Status: 1}
	readRepo.types["id-2"] = &appRepo.DictTypeDTO{ID: "id-2", Code: "nation", Name: "民族", Status: 1}
	readRepo.types["id-3"] = &appRepo.DictTypeDTO{ID: "id-3", Code: "region", Name: "地区", Status: 0}

	t.Run("list all", func(t *testing.T) {
		result, err := svc.List(context.Background(), 1, 20, -1)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result.Total != 3 {
			t.Errorf("expected total 3, got %d", result.Total)
		}
		if result.Page != 1 {
			t.Errorf("expected page 1, got %d", result.Page)
		}
	})

	t.Run("filter by status enabled", func(t *testing.T) {
		result, err := svc.List(context.Background(), 1, 20, 1)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result.Total != 2 {
			t.Errorf("expected total 2 (enabled), got %d", result.Total)
		}
	})

	t.Run("filter by status disabled", func(t *testing.T) {
		result, err := svc.List(context.Background(), 1, 20, 0)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result.Total != 1 {
			t.Errorf("expected total 1 (disabled), got %d", result.Total)
		}
	})

	t.Run("pagination", func(t *testing.T) {
		result, err := svc.List(context.Background(), 1, 1, -1)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result.Total != 3 {
			t.Errorf("expected total 3, got %d", result.Total)
		}
		if list, ok := result.List.([]*appRepo.DictTypeDTO); !ok || len(list) != 1 {
			t.Errorf("expected 1 item in list, got %v", result.List)
		}
	})
}

func TestDictQueryService_GetEntriesByTypeID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, readRepo, _ := newQueryTestService()
		readRepo.entries["e-1"] = &appRepo.DictEntryDTO{
			ID: "e-1", TypeID: "t-1", Label: "男", Value: "male",
		}
		readRepo.entries["e-2"] = &appRepo.DictEntryDTO{
			ID: "e-2", TypeID: "t-1", Label: "女", Value: "female",
		}

		entries, err := svc.GetEntriesByTypeID(context.Background(), "t-1")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(entries) != 2 {
			t.Errorf("expected 2 entries, got %d", len(entries))
		}
	})

	t.Run("empty list", func(t *testing.T) {
		svc, _, _ := newQueryTestService()
		entries, err := svc.GetEntriesByTypeID(context.Background(), "nonexistent")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(entries) != 0 {
			t.Errorf("expected empty list, got %d entries", len(entries))
		}
	})
}

func TestDictQueryService_GetEntriesByCode(t *testing.T) {
	t.Run("cache hit", func(t *testing.T) {
		svc, _, cacheRepo := newQueryTestService()
		cacheRepo.cache["gender"] = []*appRepo.DictEntryEntry{
			{Label: "男", Value: "male"},
		}

		entries, err := svc.GetEntriesByCode(context.Background(), "gender")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(entries) != 1 {
			t.Errorf("expected 1 entry, got %d", len(entries))
		}
		if entries[0].Label != "男" {
			t.Errorf("expected label '男', got %q", entries[0].Label)
		}
	})

	t.Run("cache miss", func(t *testing.T) {
		svc, _, _ := newQueryTestService()

		entries, err := svc.GetEntriesByCode(context.Background(), "nonexistent")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(entries) != 0 {
			t.Errorf("expected empty list, got %d entries", len(entries))
		}
	})
}

func TestDictQueryService_GetEntryByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, readRepo, _ := newQueryTestService()
		readRepo.entries["e-1"] = &appRepo.DictEntryDTO{
			ID: "e-1", TypeID: "t-1", Label: "男", Value: "male",
		}

		entry, err := svc.GetEntryByID(context.Background(), "e-1")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if entry.Label != "男" {
			t.Errorf("expected label '男', got %q", entry.Label)
		}
	})

	t.Run("not found", func(t *testing.T) {
		svc, _, _ := newQueryTestService()
		_, err := svc.GetEntryByID(context.Background(), "nonexistent")
		if err != appErrors.ErrNotFound {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}