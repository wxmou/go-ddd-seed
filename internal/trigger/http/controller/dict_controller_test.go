package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-ddd-seed/go-ddd-seed/internal/application/commandHandler"
	"github.com/go-ddd-seed/go-ddd-seed/internal/application/queryService"
	appRepo "github.com/go-ddd-seed/go-ddd-seed/internal/application/port/repo"
	"github.com/go-ddd-seed/go-ddd-seed/internal/domain/model/dict_type"

	appErrors "github.com/go-ddd-seed/go-ddd-seed/pkg/errors"
)

// ---------- shared mocks (reuses same pattern as commandHandler/queryService tests) ----------

type ctrlMockDictRepo struct {
	types map[string]*dict_type.DictType
}

func newCtrlMockDictRepo() *ctrlMockDictRepo {
	return &ctrlMockDictRepo{types: make(map[string]*dict_type.DictType)}
}

func (m *ctrlMockDictRepo) FindByID(_ context.Context, id string) (*dict_type.DictType, error) {
	dt, ok := m.types[id]
	if !ok {
		return nil, appErrors.ErrNotFound
	}
	return dt, nil
}

func (m *ctrlMockDictRepo) FindByCode(_ context.Context, code string) (*dict_type.DictType, error) {
	for _, dt := range m.types {
		if dt.Code == code {
			return dt, nil
		}
	}
	return nil, nil
}

func (m *ctrlMockDictRepo) FindEntryTypeID(_ context.Context, entryID string) (string, error) {
	for _, dt := range m.types {
		for _, e := range dt.Entries {
			if e.ID == entryID {
				return dt.ID, nil
			}
		}
	}
	return "", appErrors.ErrNotFound
}

func (m *ctrlMockDictRepo) Save(_ context.Context, dt *dict_type.DictType) error {
	m.types[dt.ID] = dt
	return nil
}

func (m *ctrlMockDictRepo) Delete(_ context.Context, id string) error {
	if _, ok := m.types[id]; !ok {
		return appErrors.ErrNotFound
	}
	delete(m.types, id)
	return nil
}

type ctrlMockReadRepo struct {
	types   map[string]*appRepo.DictTypeDTO
	entries map[string]*appRepo.DictEntryDTO
}

func newCtrlMockReadRepo() *ctrlMockReadRepo {
	return &ctrlMockReadRepo{
		types:   make(map[string]*appRepo.DictTypeDTO),
		entries: make(map[string]*appRepo.DictEntryDTO),
	}
}

func (m *ctrlMockReadRepo) FindByID(_ context.Context, id string) (*appRepo.DictTypeDTO, error) {
	dto, ok := m.types[id]
	if !ok {
		return nil, appErrors.ErrNotFound
	}
	return dto, nil
}

func (m *ctrlMockReadRepo) FindByCode(_ context.Context, code string) (*appRepo.DictTypeDTO, error) {
	for _, dto := range m.types {
		if dto.Code == code {
			return dto, nil
		}
	}
	return nil, appErrors.ErrNotFound
}

func (m *ctrlMockReadRepo) List(_ context.Context, offset, limit int, status int) ([]*appRepo.DictTypeDTO, int64, error) {
	var filtered []*appRepo.DictTypeDTO
	for _, dto := range m.types {
		if status < 0 || dto.Status == status {
			filtered = append(filtered, dto)
		}
	}
	total := int64(len(filtered))
	// simple offset/limit
	if offset >= len(filtered) {
		return []*appRepo.DictTypeDTO{}, total, nil
	}
	end := offset + limit
	if end > len(filtered) {
		end = len(filtered)
	}
	return filtered[offset:end], total, nil
}

func (m *ctrlMockReadRepo) FindEntriesByTypeID(_ context.Context, typeID string) ([]*appRepo.DictEntryDTO, error) {
	var result []*appRepo.DictEntryDTO
	for _, e := range m.entries {
		if e.TypeID == typeID {
			result = append(result, e)
		}
	}
	return result, nil
}

func (m *ctrlMockReadRepo) FindEntriesByTypeCode(_ context.Context, code string) ([]*appRepo.DictEntryDTO, error) {
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

func (m *ctrlMockReadRepo) FindEntryByID(_ context.Context, entryID string) (*appRepo.DictEntryDTO, error) {
	e, ok := m.entries[entryID]
	if !ok {
		return nil, appErrors.ErrNotFound
	}
	return e, nil
}

type ctrlMockCacheRepo struct {
	cache        map[string][]*appRepo.DictEntryEntry
	refreshCalls []string
	deleteCalls  []string
}

func newCtrlMockCacheRepo() *ctrlMockCacheRepo {
	return &ctrlMockCacheRepo{cache: make(map[string][]*appRepo.DictEntryEntry)}
}

func (m *ctrlMockCacheRepo) WarmUp(_ context.Context) error { return nil }
func (m *ctrlMockCacheRepo) GetEntriesByCode(_ context.Context, code string) ([]*appRepo.DictEntryEntry, error) {
	entries, ok := m.cache[code]
	if !ok {
		return nil, nil
	}
	return entries, nil
}

func (m *ctrlMockCacheRepo) RefreshByCode(_ context.Context, code string, entries []*appRepo.DictEntryEntry) error {
	m.refreshCalls = append(m.refreshCalls, code)
	if len(entries) == 0 {
		delete(m.cache, code)
		return nil
	}
	m.cache[code] = entries
	return nil
}

func (m *ctrlMockCacheRepo) DeleteByCode(_ context.Context, code string) error {
	m.deleteCalls = append(m.deleteCalls, code)
	delete(m.cache, code)
	return nil
}

// ---------- test setup ----------

type ctrlTestEnv struct {
	engine    *gin.Engine
	dictRepo  *ctrlMockDictRepo
	readRepo  *ctrlMockReadRepo
	cacheRepo *ctrlMockCacheRepo
}

func newCtrlTestEnv() *ctrlTestEnv {
	gin.SetMode(gin.TestMode)

	dictRepo := newCtrlMockDictRepo()
	readRepo := newCtrlMockReadRepo()
	cacheRepo := newCtrlMockCacheRepo()

	handler := commandHandler.NewDictCommandHandler(dictRepo, cacheRepo)
	svc := queryService.NewDictQueryService(readRepo, cacheRepo)
	ctrl := NewDictController(handler, svc)

	engine := gin.New()
	api := engine.Group("/api/v1")
	ctrl.RegisterRoutes(api)

	return &ctrlTestEnv{
		engine:    engine,
		dictRepo:  dictRepo,
		readRepo:  readRepo,
		cacheRepo: cacheRepo,
	}
}

func (env *ctrlTestEnv) doRequest(method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.engine.ServeHTTP(w, req)
	return w
}

func seedDictType(env *ctrlTestEnv) (*dict_type.DictType, *appRepo.DictTypeDTO) {
	dt, _ := dict_type.NewDictType("id-1", "gender", "性别", "性别字典")
	env.dictRepo.types["id-1"] = dt

	readDTO := &appRepo.DictTypeDTO{
		ID: "id-1", Code: "gender", Name: "性别", Description: "性别字典", Status: 1,
	}
	env.readRepo.types["id-1"] = readDTO

	return dt, readDTO
}

type apiResp struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

func parseResp(t *testing.T, w *httptest.ResponseRecorder) *apiResp {
	t.Helper()
	var r apiResp
	if err := json.Unmarshal(w.Body.Bytes(), &r); err != nil {
		t.Fatalf("failed to parse response: %v, body: %s", err, w.Body.String())
	}
	return &r
}

// ---------- DictType endpoints ----------

func TestDictController_CreateType(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := newCtrlTestEnv()
		w := env.doRequest("POST", "/api/v1/infra/dict-types",
			`{"code":"gender","name":"性别","description":"性别字典"}`)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
		r := parseResp(t, w)
		if r.Code != 0 {
			t.Errorf("expected code 0, got %d", r.Code)
		}
	})

	t.Run("bad request with missing fields", func(t *testing.T) {
		env := newCtrlTestEnv()
		w := env.doRequest("POST", "/api/v1/infra/dict-types", `{}`)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("duplicate code", func(t *testing.T) {
		env := newCtrlTestEnv()
		seedDictType(env)

		w := env.doRequest("POST", "/api/v1/infra/dict-types",
			`{"code":"gender","name":"性别2","description":""}`)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestDictController_UpdateType(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := newCtrlTestEnv()
		seedDictType(env)

		w := env.doRequest("PUT", "/api/v1/infra/dict-types/id-1",
			`{"name":"性别2","description":"新描述"}`)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("not found", func(t *testing.T) {
		env := newCtrlTestEnv()
		w := env.doRequest("PUT", "/api/v1/infra/dict-types/nonexistent",
			`{"name":"test","description":""}`)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", w.Code)
		}
	})
}

func TestDictController_DeleteType(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := newCtrlTestEnv()
		seedDictType(env)

		w := env.doRequest("DELETE", "/api/v1/infra/dict-types/id-1", "")

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestDictController_EnableType(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := newCtrlTestEnv()
		dt, _ := seedDictType(env)
		_ = dt.Disable()

		w := env.doRequest("PUT", "/api/v1/infra/dict-types/id-1/enable", "")

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("already enabled", func(t *testing.T) {
		env := newCtrlTestEnv()
		seedDictType(env)

		w := env.doRequest("PUT", "/api/v1/infra/dict-types/id-1/enable", "")

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestDictController_DisableType(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := newCtrlTestEnv()
		seedDictType(env)

		w := env.doRequest("PUT", "/api/v1/infra/dict-types/id-1/disable", "")

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("already disabled", func(t *testing.T) {
		env := newCtrlTestEnv()
		dt, _ := seedDictType(env)
		_ = dt.Disable()

		w := env.doRequest("PUT", "/api/v1/infra/dict-types/id-1/disable", "")

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestDictController_GetTypeByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := newCtrlTestEnv()
		seedDictType(env)

		w := env.doRequest("GET", "/api/v1/infra/dict-types/id-1", "")

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
		r := parseResp(t, w)
		if r.Code != 0 {
			t.Errorf("expected code 0, got %d", r.Code)
		}
	})

	t.Run("not found", func(t *testing.T) {
		env := newCtrlTestEnv()
		w := env.doRequest("GET", "/api/v1/infra/dict-types/nonexistent", "")

		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", w.Code)
		}
	})
}

func TestDictController_GetTypeByCode(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := newCtrlTestEnv()
		seedDictType(env)

		w := env.doRequest("GET", "/api/v1/infra/dict-types/code/gender", "")

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
		r := parseResp(t, w)
		if r.Code != 0 {
			t.Errorf("expected code 0, got %d", r.Code)
		}
	})
}

func TestDictController_ListTypes(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := newCtrlTestEnv()
		seedDictType(env)

		w := env.doRequest("GET", "/api/v1/infra/dict-types?page=1&page_size=20", "")

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("with status filter", func(t *testing.T) {
		env := newCtrlTestEnv()
		seedDictType(env)

		w := env.doRequest("GET", "/api/v1/infra/dict-types?page=1&page_size=20&status=1", "")

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
}

// ---------- DictEntry endpoints ----------

func seedDictTypeWithEntry(env *ctrlTestEnv) {
	dt, _ := seedDictType(env)
	entry := &dict_type.DictEntry{
		ID: "e-1", Label: "男", Value: "male", SortOrder: 1, Status: dict_type.DictEntryStatusEnabled,
	}
	_ = dt.AddEntry(entry)
	env.readRepo.entries["e-1"] = &appRepo.DictEntryDTO{
		ID: "e-1", TypeID: "id-1", Label: "男", Value: "male", SortOrder: 1, Status: 1,
	}
}

func TestDictController_AddEntry(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := newCtrlTestEnv()
		seedDictType(env)

		w := env.doRequest("POST", "/api/v1/infra/dict-entries",
			`{"type_id":"id-1","label":"男","value":"male","sort_order":1}`)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("bad request with missing fields", func(t *testing.T) {
		env := newCtrlTestEnv()
		w := env.doRequest("POST", "/api/v1/infra/dict-entries", `{}`)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
}

func TestDictController_UpdateEntry(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := newCtrlTestEnv()
		seedDictTypeWithEntry(env)

		w := env.doRequest("PUT", "/api/v1/infra/dict-entries/e-1",
			`{"label":"男性","value":"male","sort_order":2}`)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("not found", func(t *testing.T) {
		env := newCtrlTestEnv()
		w := env.doRequest("PUT", "/api/v1/infra/dict-entries/nonexistent",
			`{"label":"test","value":"test","sort_order":1}`)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", w.Code)
		}
	})
}

func TestDictController_RemoveEntry(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := newCtrlTestEnv()
		seedDictTypeWithEntry(env)

		w := env.doRequest("DELETE", "/api/v1/infra/dict-entries/e-1", "")

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestDictController_EnableEntry(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := newCtrlTestEnv()
		seedDictTypeWithEntry(env)
		// disable the entry first
		_ = env.dictRepo.types["id-1"].DisableEntry("e-1")

		w := env.doRequest("PUT", "/api/v1/infra/dict-entries/e-1/enable", "")

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestDictController_DisableEntry(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := newCtrlTestEnv()
		seedDictTypeWithEntry(env)

		w := env.doRequest("PUT", "/api/v1/infra/dict-entries/e-1/disable", "")

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestDictController_GetEntryByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := newCtrlTestEnv()
		seedDictTypeWithEntry(env)

		w := env.doRequest("GET", "/api/v1/infra/dict-entries/e-1", "")

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
}

func TestDictController_ListEntries(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := newCtrlTestEnv()
		seedDictTypeWithEntry(env)

		w := env.doRequest("GET", "/api/v1/infra/dict-entries?type_id=id-1", "")

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("missing type_id", func(t *testing.T) {
		env := newCtrlTestEnv()
		w := env.doRequest("GET", "/api/v1/infra/dict-entries", "")

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
}

func TestDictController_GetEntriesByCode(t *testing.T) {
	t.Run("success from cache", func(t *testing.T) {
		env := newCtrlTestEnv()
		seedDictType(env)
		env.cacheRepo.cache["gender"] = []*appRepo.DictEntryEntry{
			{Label: "男", Value: "male"},
		}

		w := env.doRequest("GET", "/api/v1/infra/dicts/entries/gender", "")

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
		r := parseResp(t, w)
		if r.Code != 0 {
			t.Errorf("expected code 0, got %d: %s", r.Code, w.Body.String())
		}
	})

	t.Run("cache miss, DB fallback", func(t *testing.T) {
		env := newCtrlTestEnv()
		seedDictType(env)
		env.readRepo.entries["e-1"] = &appRepo.DictEntryDTO{
			ID: "e-1", TypeID: "id-1", Label: "男", Value: "male", Status: 1,
		}

		w := env.doRequest("GET", "/api/v1/infra/dicts/entries/gender", "")

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})
}