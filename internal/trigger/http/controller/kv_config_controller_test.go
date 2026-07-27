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
	"github.com/go-ddd-seed/go-ddd-seed/internal/domain/model/kv_config"
	appErrors "github.com/go-ddd-seed/go-ddd-seed/pkg/errors"
)

// ---------- mocks ----------

type kvCtrlMockRepo struct {
	configs map[string]*kv_config.KvConfig
}

func newKvCtrlMockRepo() *kvCtrlMockRepo {
	return &kvCtrlMockRepo{configs: make(map[string]*kv_config.KvConfig)}
}

func (m *kvCtrlMockRepo) Save(_ context.Context, config *kv_config.KvConfig) error {
	m.configs[config.ID] = config
	return nil
}

func (m *kvCtrlMockRepo) FindByID(_ context.Context, id string) (*kv_config.KvConfig, error) {
	config, ok := m.configs[id]
	if !ok {
		return nil, appErrors.ErrNotFound
	}
	return config, nil
}

func (m *kvCtrlMockRepo) FindByKey(_ context.Context, key string) (*kv_config.KvConfig, error) {
	for _, config := range m.configs {
		if config.Key == key {
			return config, nil
		}
	}
	return nil, nil
}

func (m *kvCtrlMockRepo) Delete(_ context.Context, id string) error {
	if _, ok := m.configs[id]; !ok {
		return appErrors.ErrNotFound
	}
	delete(m.configs, id)
	return nil
}

type kvCtrlMockReadRepo struct {
	configs map[string]*appRepo.KvConfigDTO
}

func newKvCtrlMockReadRepo() *kvCtrlMockReadRepo {
	return &kvCtrlMockReadRepo{configs: make(map[string]*appRepo.KvConfigDTO)}
}

func (m *kvCtrlMockReadRepo) FindByID(_ context.Context, id string) (*appRepo.KvConfigDTO, error) {
	dto, ok := m.configs[id]
	if !ok {
		return nil, appErrors.ErrNotFound
	}
	return dto, nil
}

func (m *kvCtrlMockReadRepo) FindByKey(_ context.Context, key string) (*appRepo.KvConfigDTO, error) {
	for _, dto := range m.configs {
		if dto.Key == key {
			return dto, nil
		}
	}
	return nil, appErrors.ErrNotFound
}

func (m *kvCtrlMockReadRepo) List(_ context.Context, offset, limit int, status int) ([]*appRepo.KvConfigDTO, int64, error) {
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

// ---------- test setup ----------

type kvCtrlTestEnv struct {
	engine   *gin.Engine
	repo     *kvCtrlMockRepo
	readRepo *kvCtrlMockReadRepo
}

func newKvCtrlTestEnv() *kvCtrlTestEnv {
	gin.SetMode(gin.TestMode)

	repo := newKvCtrlMockRepo()
	readRepo := newKvCtrlMockReadRepo()

	handler := commandHandler.NewKvConfigCommandHandler(repo)
	svc := queryService.NewKvConfigQueryService(readRepo)
	ctrl := NewKvConfigController(handler, svc)

	engine := gin.New()
	api := engine.Group("/api/v1")
	ctrl.RegisterRoutes(api)

	return &kvCtrlTestEnv{
		engine:   engine,
		repo:     repo,
		readRepo: readRepo,
	}
}

func (env *kvCtrlTestEnv) doRequest(method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.engine.ServeHTTP(w, req)
	return w
}

func seedKvConfig(env *kvCtrlTestEnv) *kv_config.KvConfig {
	kv, _ := kv_config.NewKvConfig("id-1", "site_name", "TestApp", "站点名称")
	env.repo.configs["id-1"] = kv
	env.readRepo.configs["id-1"] = &appRepo.KvConfigDTO{
		ID: "id-1", Key: "site_name", Value: "TestApp", Description: "站点名称", Status: 1,
	}
	return kv
}

type kvAPIResp struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

func parseKVResp(t *testing.T, w *httptest.ResponseRecorder) *kvAPIResp {
	t.Helper()
	var r kvAPIResp
	if err := json.Unmarshal(w.Body.Bytes(), &r); err != nil {
		t.Fatalf("failed to parse response: %v, body: %s", err, w.Body.String())
	}
	return &r
}

// ---------- tests ----------

func TestKvConfigController_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := newKvCtrlTestEnv()
		w := env.doRequest("POST", "/api/v1/infra/kv-configs",
			`{"key":"site_name","value":"TestApp","description":"站点名称"}`)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		r := parseKVResp(t, w)
		if r.Code != 0 {
			t.Errorf("expected code 0, got %d", r.Code)
		}
	})

	t.Run("bad request", func(t *testing.T) {
		env := newKvCtrlTestEnv()
		w := env.doRequest("POST", "/api/v1/infra/kv-configs", `{}`)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("duplicate key", func(t *testing.T) {
		env := newKvCtrlTestEnv()
		seedKvConfig(env)

		w := env.doRequest("POST", "/api/v1/infra/kv-configs",
			`{"key":"site_name","value":"New"}`)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestKvConfigController_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := newKvCtrlTestEnv()
		seedKvConfig(env)

		w := env.doRequest("PUT", "/api/v1/infra/kv-configs/id-1",
			`{"value":"NewName","description":"新描述"}`)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("not found", func(t *testing.T) {
		env := newKvCtrlTestEnv()
		w := env.doRequest("PUT", "/api/v1/infra/kv-configs/nonexistent",
			`{"value":"test","description":""}`)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", w.Code)
		}
	})
}

func TestKvConfigController_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := newKvCtrlTestEnv()
		seedKvConfig(env)

		w := env.doRequest("DELETE", "/api/v1/infra/kv-configs/id-1", "")

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestKvConfigController_GetByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := newKvCtrlTestEnv()
		seedKvConfig(env)

		w := env.doRequest("GET", "/api/v1/infra/kv-configs/id-1", "")

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
		r := parseKVResp(t, w)
		if r.Code != 0 {
			t.Errorf("expected code 0, got %d", r.Code)
		}
	})

	t.Run("not found", func(t *testing.T) {
		env := newKvCtrlTestEnv()
		w := env.doRequest("GET", "/api/v1/infra/kv-configs/nonexistent", "")

		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", w.Code)
		}
	})
}

func TestKvConfigController_GetByKey(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := newKvCtrlTestEnv()
		seedKvConfig(env)

		w := env.doRequest("GET", "/api/v1/infra/kv-configs/key/site_name", "")

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
}

func TestKvConfigController_List(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := newKvCtrlTestEnv()
		seedKvConfig(env)

		w := env.doRequest("GET", "/api/v1/infra/kv-configs?page=1&page_size=20", "")

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("with status filter", func(t *testing.T) {
		env := newKvCtrlTestEnv()
		seedKvConfig(env)

		w := env.doRequest("GET", "/api/v1/infra/kv-configs?page=1&page_size=20&status=1", "")

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
}