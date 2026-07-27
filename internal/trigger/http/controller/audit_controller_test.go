package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	appRepo "github.com/go-ddd-seed/go-ddd-seed/internal/application/port/repo"
	"github.com/go-ddd-seed/go-ddd-seed/internal/application/queryService"
)

// mockAuditQuerySvcReadRepo 审计日志查询服务使用的 mock 读仓储
type mockAuditQuerySvcReadRepo struct {
	dtos []*appRepo.AuditLogReadDTO
}

func (m *mockAuditQuerySvcReadRepo) List(_ context.Context, query appRepo.AuditLogQuery) ([]*appRepo.AuditLogReadDTO, int64, error) {
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

// auditCtrlTestEnv 审计控制器测试环境
type auditCtrlTestEnv struct {
	engine *gin.Engine
}

func newAuditCtrlTestEnv(dtos []*appRepo.AuditLogReadDTO) *auditCtrlTestEnv {
	gin.SetMode(gin.TestMode)

	readRepo := &mockAuditQuerySvcReadRepo{dtos: dtos}
	svc := queryService.NewAuditLogQueryService(readRepo)
	ctrl := NewAuditController(svc)

	engine := gin.New()
	api := engine.Group("/api/v1")
	ctrl.RegisterRoutes(api)

	return &auditCtrlTestEnv{engine: engine}
}

func (env *auditCtrlTestEnv) doRequest(path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("GET", path, nil)
	w := httptest.NewRecorder()
	env.engine.ServeHTTP(w, req)
	return w
}

type auditAPIResp struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

func parseAuditResp(t *testing.T, w *httptest.ResponseRecorder) *auditAPIResp {
	t.Helper()
	var r auditAPIResp
	if err := json.Unmarshal(w.Body.Bytes(), &r); err != nil {
		t.Fatalf("failed to parse response: %v, body: %s", err, w.Body.String())
	}
	return &r
}

func TestAuditController_List(t *testing.T) {
	now := time.Now()

	t.Run("success", func(t *testing.T) {
		env := newAuditCtrlTestEnv([]*appRepo.AuditLogReadDTO{
			{ID: "1", OperatorID: "u1", OperatorName: "admin", Action: "create",
				TargetType: "role", TargetID: "r1", ClientIP: "127.0.0.1",
				UserAgent: "test", TraceID: "trace-1", CreatedAt: now},
		})

		w := env.doRequest("/api/v1/audit-logs?page=1&page_size=20")

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
		}

		r := parseAuditResp(t, w)
		if r.Code != 0 {
			t.Errorf("expected code 0, got %d", r.Code)
		}
	})

	t.Run("empty result", func(t *testing.T) {
		env := newAuditCtrlTestEnv(nil)

		w := env.doRequest("/api/v1/audit-logs?page=1&page_size=20")

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}

		r := parseAuditResp(t, w)
		if r.Code != 0 {
			t.Errorf("expected code 0, got %d", r.Code)
		}
	})

	t.Run("filter by operator_id", func(t *testing.T) {
		env := newAuditCtrlTestEnv([]*appRepo.AuditLogReadDTO{
			{ID: "1", OperatorID: "u1", Action: "create", TargetType: "role", TargetID: "r1", CreatedAt: now},
			{ID: "2", OperatorID: "u2", Action: "update", TargetType: "user", TargetID: "u1", CreatedAt: now},
		})

		w := env.doRequest("/api/v1/audit-logs?operator_id=u1&page=1&page_size=20")

		var respData struct {
			Items []struct {
				ID string `json:"id"`
			} `json:"items"`
			Total int64 `json:"total"`
			Page  int   `json:"page"`
		}
		if err := json.Unmarshal(parseAuditResp(t, w).Data, &respData); err != nil {
			t.Fatalf("failed to parse data: %v", err)
		}
		if respData.Total != 1 {
			t.Errorf("expected total 1, got %d", respData.Total)
		}
		if len(respData.Items) != 1 || respData.Items[0].ID != "1" {
			t.Errorf("expected item 1, got %+v", respData.Items)
		}
	})

	t.Run("filter by action", func(t *testing.T) {
		env := newAuditCtrlTestEnv([]*appRepo.AuditLogReadDTO{
			{ID: "1", OperatorID: "u1", Action: "create", TargetType: "role", TargetID: "r1", CreatedAt: now},
			{ID: "2", OperatorID: "u1", Action: "delete", TargetType: "role", TargetID: "r2", CreatedAt: now},
		})

		w := env.doRequest("/api/v1/audit-logs?action=delete&page=1&page_size=20")

		var respData struct {
			Items []struct {
				ID     string `json:"id"`
				Action string `json:"action"`
			} `json:"items"`
			Total int64 `json:"total"`
		}
		if err := json.Unmarshal(parseAuditResp(t, w).Data, &respData); err != nil {
			t.Fatalf("failed to parse data: %v", err)
		}
		if respData.Total != 1 {
			t.Errorf("expected total 1, got %d", respData.Total)
		}
		if len(respData.Items) != 1 || respData.Items[0].Action != "delete" {
			t.Errorf("expected delete action, got %+v", respData.Items[0])
		}
	})

	t.Run("pagination", func(t *testing.T) {
		env := newAuditCtrlTestEnv([]*appRepo.AuditLogReadDTO{
			{ID: "1", OperatorID: "u1", Action: "create", TargetType: "role", TargetID: "r1", CreatedAt: now},
			{ID: "2", OperatorID: "u1", Action: "update", TargetType: "role", TargetID: "r1", CreatedAt: now},
			{ID: "3", OperatorID: "u2", Action: "delete", TargetType: "user", TargetID: "u1", CreatedAt: now},
		})

		w := env.doRequest("/api/v1/audit-logs?page=1&page_size=2")

		var respData struct {
			Items []struct {
				ID string `json:"id"`
			} `json:"items"`
			Total int64 `json:"total"`
			Page  int   `json:"page"`
		}
		if err := json.Unmarshal(parseAuditResp(t, w).Data, &respData); err != nil {
			t.Fatalf("failed to parse data: %v", err)
		}
		if respData.Total != 3 {
			t.Errorf("expected total 3, got %d", respData.Total)
		}
		if len(respData.Items) != 2 {
			t.Errorf("expected 2 items on page 1, got %d", len(respData.Items))
		}
	})

	t.Run("default pagination", func(t *testing.T) {
		env := newAuditCtrlTestEnv([]*appRepo.AuditLogReadDTO{
			{ID: "1", OperatorID: "u1", Action: "create", TargetType: "role", TargetID: "r1", CreatedAt: now},
		})

		w := env.doRequest("/api/v1/audit-logs")

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("response format", func(t *testing.T) {
		env := newAuditCtrlTestEnv([]*appRepo.AuditLogReadDTO{
			{ID: "1", OperatorID: "u1", OperatorName: "admin", Action: "create",
				TargetType: "role", TargetID: "r1", ClientIP: "127.0.0.1",
				UserAgent: "Go-http-client", TraceID: "trace-1", CreatedAt: now},
		})

		w := env.doRequest("/api/v1/audit-logs?page=1&page_size=20")

		r := parseAuditResp(t, w)
		var respData struct {
			Items []struct {
				ID           string `json:"id"`
				OperatorID   string `json:"operator_id"`
				OperatorName string `json:"operator_name"`
				Action       string `json:"action"`
				TargetType   string `json:"target_type"`
				TargetID     string `json:"target_id"`
				ClientIP     string `json:"client_ip"`
				UserAgent    string `json:"user_agent"`
				TraceID      string `json:"trace_id"`
				CreatedAt    string `json:"created_at"`
			} `json:"items"`
			Total int64 `json:"total"`
			Page  int   `json:"page"`
		}
		if err := json.Unmarshal(r.Data, &respData); err != nil {
			t.Fatalf("failed to parse data: %v", err)
		}

		if len(respData.Items) != 1 {
			t.Fatalf("expected 1 item, got %d", len(respData.Items))
		}
		item := respData.Items[0]
		if item.ID != "1" || item.OperatorID != "u1" || item.OperatorName != "admin" {
			t.Errorf("unexpected item fields: %+v", item)
		}
		if item.Action != "create" || item.TargetType != "role" || item.TargetID != "r1" {
			t.Errorf("unexpected action/target: %+v", item)
		}
		if item.CreatedAt == "" {
			t.Errorf("expected created_at to be set, got empty")
		}
	})
}