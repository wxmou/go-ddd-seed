package commandHandler

import (
	"context"
	"testing"

	appRepo "github.com/go-ddd-seed/go-ddd-seed/internal/application/port/repo"
	"github.com/go-ddd-seed/go-ddd-seed/internal/application/command"
	"github.com/go-ddd-seed/go-ddd-seed/internal/domain/model/dict_type"

	appErrors "github.com/go-ddd-seed/go-ddd-seed/pkg/errors"
)

// ---------- mocks ----------

type mockDictRepo struct {
	types map[string]*dict_type.DictType
}

func newMockDictRepo() *mockDictRepo {
	return &mockDictRepo{types: make(map[string]*dict_type.DictType)}
}

func (m *mockDictRepo) FindByID(_ context.Context, id string) (*dict_type.DictType, error) {
	dt, ok := m.types[id]
	if !ok {
		return nil, appErrors.ErrNotFound
	}
	return dt, nil
}

func (m *mockDictRepo) FindByCode(_ context.Context, code string) (*dict_type.DictType, error) {
	for _, dt := range m.types {
		if dt.Code == code {
			return dt, nil
		}
	}
	return nil, nil
}

func (m *mockDictRepo) FindEntryTypeID(_ context.Context, entryID string) (string, error) {
	for _, dt := range m.types {
		for _, e := range dt.Entries {
			if e.ID == entryID {
				return dt.ID, nil
			}
		}
	}
	return "", appErrors.ErrNotFound
}

func (m *mockDictRepo) Save(_ context.Context, dt *dict_type.DictType) error {
	m.types[dt.ID] = dt
	return nil
}

func (m *mockDictRepo) Delete(_ context.Context, id string) error {
	if _, ok := m.types[id]; !ok {
		return appErrors.ErrNotFound
	}
	delete(m.types, id)
	return nil
}

type mockCacheRepo struct {
	cache        map[string][]*appRepo.DictEntryEntry
	refreshCalls []string
	deleteCalls  []string
}

func newMockCacheRepo() *mockCacheRepo {
	return &mockCacheRepo{cache: make(map[string][]*appRepo.DictEntryEntry)}
}

func (m *mockCacheRepo) WarmUp(_ context.Context) error {
	return nil
}

func (m *mockCacheRepo) GetEntriesByCode(_ context.Context, code string) ([]*appRepo.DictEntryEntry, error) {
	entries, ok := m.cache[code]
	if !ok {
		return nil, nil
	}
	return entries, nil
}

func (m *mockCacheRepo) RefreshByCode(_ context.Context, code string, entries []*appRepo.DictEntryEntry) error {
	m.refreshCalls = append(m.refreshCalls, code)
	if len(entries) == 0 {
		delete(m.cache, code)
		return nil
	}
	m.cache[code] = entries
	return nil
}

func (m *mockCacheRepo) DeleteByCode(_ context.Context, code string) error {
	m.deleteCalls = append(m.deleteCalls, code)
	delete(m.cache, code)
	return nil
}

// ---------- helpers ----------

func newTestHandler() (*DictCommandHandler, *mockDictRepo, *mockCacheRepo) {
	dictRepo := newMockDictRepo()
	cacheRepo := newMockCacheRepo()
	handler := NewDictCommandHandler(dictRepo, cacheRepo)
	return handler, dictRepo, cacheRepo
}

// ---------- DictType ----------

func TestDictCommandHandler_CreateType(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		handler, _, _ := newTestHandler()

		result, err := handler.CreateType(context.Background(), &command.CreateDictTypeCommand{
			Code: "gender", Name: "性别", Description: "性别字典",
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result.Code != "gender" {
			t.Errorf("expected code 'gender', got %q", result.Code)
		}
		if result.Name != "性别" {
			t.Errorf("expected name '性别', got %q", result.Name)
		}
	})

	t.Run("duplicate code", func(t *testing.T) {
		handler, dictRepo, _ := newTestHandler()

		// pre-seed dictRepo with existing type
		dt, _ := dict_type.NewDictType("existing", "gender", "性别", "")
		dictRepo.types["existing"] = dt

		_, err := handler.CreateType(context.Background(), &command.CreateDictTypeCommand{
			Code: "gender", Name: "性别",
		})
		if err != ErrDictTypeCodeDuplicate {
			t.Errorf("expected ErrDictTypeCodeDuplicate, got %v", err)
		}
	})

	t.Run("empty code from domain validation", func(t *testing.T) {
		handler, _, _ := newTestHandler()
		_, err := handler.CreateType(context.Background(), &command.CreateDictTypeCommand{
			Code: "", Name: "test",
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestDictCommandHandler_UpdateType(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		handler, dictRepo, _ := newTestHandler()
		dt, _ := dict_type.NewDictType("id-1", "gender", "性别", "")
		dictRepo.types["id-1"] = dt

		result, err := handler.UpdateType(context.Background(), &command.UpdateDictTypeCommand{
			ID: "id-1", Name: "性别2", Description: "新描述",
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result.Name != "性别2" {
			t.Errorf("expected name '性别2', got %q", result.Name)
		}
		if result.Description != "新描述" {
			t.Errorf("expected description '新描述', got %q", result.Description)
		}
	})

	t.Run("not found", func(t *testing.T) {
		handler, _, _ := newTestHandler()
		_, err := handler.UpdateType(context.Background(), &command.UpdateDictTypeCommand{
			ID: "nonexistent", Name: "test",
		})
		if err != appErrors.ErrNotFound {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestDictCommandHandler_DeleteType(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		handler, dictRepo, cacheRepo := newTestHandler()
		dt, _ := dict_type.NewDictType("id-1", "gender", "性别", "")
		dictRepo.types["id-1"] = dt

		err := handler.DeleteType(context.Background(), &command.DeleteDictTypeCommand{ID: "id-1"})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// verify deleted from repo
		if _, ok := dictRepo.types["id-1"]; ok {
			t.Error("expected type to be deleted from repo")
		}

		// verify cache cleared
		if len(cacheRepo.deleteCalls) != 1 || cacheRepo.deleteCalls[0] != "gender" {
			t.Errorf("expected cacheRepo.DeleteByCode('gender'), got %v", cacheRepo.deleteCalls)
		}
	})

	t.Run("not found", func(t *testing.T) {
		handler, _, _ := newTestHandler()
		err := handler.DeleteType(context.Background(), &command.DeleteDictTypeCommand{ID: "nonexistent"})
		if err != appErrors.ErrNotFound {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestDictCommandHandler_EnableType(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		handler, dictRepo, _ := newTestHandler()
		dt, _ := dict_type.NewDictType("id-1", "gender", "性别", "")
		_ = dt.Disable()
		dictRepo.types["id-1"] = dt

		result, err := handler.EnableType(context.Background(), &command.EnableDictTypeCommand{ID: "id-1"})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !result.IsEnabled() {
			t.Error("expected type to be enabled")
		}
	})

	t.Run("already enabled", func(t *testing.T) {
		handler, dictRepo, _ := newTestHandler()
		dt, _ := dict_type.NewDictType("id-1", "gender", "性别", "")
		dictRepo.types["id-1"] = dt

		_, err := handler.EnableType(context.Background(), &command.EnableDictTypeCommand{ID: "id-1"})
		if err != dict_type.ErrDictTypeAlreadyEnabled {
			t.Errorf("expected ErrDictTypeAlreadyEnabled, got %v", err)
		}
	})
}

func TestDictCommandHandler_DisableType(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		handler, dictRepo, _ := newTestHandler()
		dt, _ := dict_type.NewDictType("id-1", "gender", "性别", "")
		dictRepo.types["id-1"] = dt

		result, err := handler.DisableType(context.Background(), &command.DisableDictTypeCommand{ID: "id-1"})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result.IsEnabled() {
			t.Error("expected type to be disabled")
		}
	})

	t.Run("already disabled", func(t *testing.T) {
		handler, dictRepo, _ := newTestHandler()
		dt, _ := dict_type.NewDictType("id-1", "gender", "性别", "")
		_ = dt.Disable()
		dictRepo.types["id-1"] = dt

		_, err := handler.DisableType(context.Background(), &command.DisableDictTypeCommand{ID: "id-1"})
		if err != dict_type.ErrDictTypeAlreadyDisabled {
			t.Errorf("expected ErrDictTypeAlreadyDisabled, got %v", err)
		}
	})
}

// ---------- DictEntry ----------

func TestDictCommandHandler_AddEntry(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		handler, dictRepo, cacheRepo := newTestHandler()
		dt, _ := dict_type.NewDictType("id-1", "gender", "性别", "")
		dictRepo.types["id-1"] = dt

		result, err := handler.AddEntry(context.Background(), &command.AddDictEntryCommand{
			TypeID: "id-1", Label: "男", Value: "male", SortOrder: 1,
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(result.Entries) != 1 {
			t.Fatalf("expected 1 entry, got %d", len(result.Entries))
		}
		if result.Entries[0].Label != "男" {
			t.Errorf("expected label '男', got %q", result.Entries[0].Label)
		}

		// verify cache was refreshed
		if len(cacheRepo.refreshCalls) != 1 || cacheRepo.refreshCalls[0] != "gender" {
			t.Errorf("expected cacheRepo.RefreshByCode('gender'), got %v", cacheRepo.refreshCalls)
		}
	})

	t.Run("type not found", func(t *testing.T) {
		handler, _, _ := newTestHandler()
		_, err := handler.AddEntry(context.Background(), &command.AddDictEntryCommand{
			TypeID: "nonexistent", Label: "男", Value: "male",
		})
		if err != appErrors.ErrNotFound {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("duplicate value", func(t *testing.T) {
		handler, dictRepo, _ := newTestHandler()
		dt, _ := dict_type.NewDictType("id-1", "gender", "性别", "")
		_ = dt.AddEntry(&dict_type.DictEntry{ID: "e-1", Label: "男", Value: "male", Status: dict_type.DictEntryStatusEnabled})
		dictRepo.types["id-1"] = dt

		_, err := handler.AddEntry(context.Background(), &command.AddDictEntryCommand{
			TypeID: "id-1", Label: "男2", Value: "male",
		})
		if err != dict_type.ErrDictEntryValueDuplicate {
			t.Errorf("expected ErrDictEntryValueDuplicate, got %v", err)
		}
	})
}

func TestDictCommandHandler_UpdateEntry(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		handler, dictRepo, cacheRepo := newTestHandler()
		dt, _ := dict_type.NewDictType("id-1", "gender", "性别", "")
		_ = dt.AddEntry(&dict_type.DictEntry{ID: "e-1", Label: "男", Value: "male", SortOrder: 1, Status: dict_type.DictEntryStatusEnabled})
		dictRepo.types["id-1"] = dt

		result, err := handler.UpdateEntry(context.Background(), &command.UpdateDictEntryCommand{
			ID: "e-1", Label: "男性", Value: "male", SortOrder: 2,
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result.Entries[0].Label != "男性" {
			t.Errorf("expected label '男性', got %q", result.Entries[0].Label)
		}

		// verify cache refresh
		if len(cacheRepo.refreshCalls) != 1 || cacheRepo.refreshCalls[0] != "gender" {
			t.Errorf("expected cacheRepo.RefreshByCode('gender'), got %v", cacheRepo.refreshCalls)
		}
	})

	t.Run("entry not found", func(t *testing.T) {
		handler, _, _ := newTestHandler()
		_, err := handler.UpdateEntry(context.Background(), &command.UpdateDictEntryCommand{
			ID: "nonexistent", Label: "test", Value: "test",
		})
		if err != appErrors.ErrNotFound {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestDictCommandHandler_RemoveEntry(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		handler, dictRepo, cacheRepo := newTestHandler()
		dt, _ := dict_type.NewDictType("id-1", "gender", "性别", "")
		_ = dt.AddEntry(&dict_type.DictEntry{ID: "e-1", Label: "男", Value: "male", Status: dict_type.DictEntryStatusEnabled})
		dictRepo.types["id-1"] = dt

		result, err := handler.RemoveEntry(context.Background(), &command.RemoveDictEntryCommand{ID: "e-1"})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(result.Entries) != 0 {
			t.Errorf("expected 0 entries, got %d", len(result.Entries))
		}

		if len(cacheRepo.refreshCalls) != 1 || cacheRepo.refreshCalls[0] != "gender" {
			t.Errorf("expected cache refresh, got %v", cacheRepo.refreshCalls)
		}
	})
}

func TestDictCommandHandler_EnableEntry(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		handler, dictRepo, cacheRepo := newTestHandler()
		dt, _ := dict_type.NewDictType("id-1", "gender", "性别", "")
		_ = dt.AddEntry(&dict_type.DictEntry{ID: "e-1", Label: "男", Value: "male", Status: dict_type.DictEntryStatusDisabled})
		dictRepo.types["id-1"] = dt

		result, err := handler.EnableEntry(context.Background(), &command.EnableDictEntryCommand{ID: "e-1"})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !result.Entries[0].IsEnabled() {
			t.Error("expected entry to be enabled")
		}

		if len(cacheRepo.refreshCalls) != 1 {
			t.Error("expected cache refresh")
		}
	})
}

func TestDictCommandHandler_DisableEntry(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		handler, dictRepo, cacheRepo := newTestHandler()
		dt, _ := dict_type.NewDictType("id-1", "gender", "性别", "")
		_ = dt.AddEntry(&dict_type.DictEntry{ID: "e-1", Label: "男", Value: "male", Status: dict_type.DictEntryStatusEnabled})
		dictRepo.types["id-1"] = dt

		result, err := handler.DisableEntry(context.Background(), &command.DisableDictEntryCommand{ID: "e-1"})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result.Entries[0].IsEnabled() {
			t.Error("expected entry to be disabled")
		}

		if len(cacheRepo.refreshCalls) != 1 {
			t.Error("expected cache refresh")
		}
	})
}