package commandHandler

import (
	"context"
	"testing"

	"github.com/go-ddd-seed/go-ddd-seed/internal/application/command"
	"github.com/go-ddd-seed/go-ddd-seed/internal/domain/model/kv_config"
	appErrors "github.com/go-ddd-seed/go-ddd-seed/pkg/errors"
)

// ---------- mocks ----------

type mockKvConfigRepo struct {
	configs map[string]*kv_config.KvConfig
}

func newMockKvConfigRepo() *mockKvConfigRepo {
	return &mockKvConfigRepo{configs: make(map[string]*kv_config.KvConfig)}
}

func (m *mockKvConfigRepo) Save(_ context.Context, config *kv_config.KvConfig) error {
	m.configs[config.ID] = config
	return nil
}

func (m *mockKvConfigRepo) FindByID(_ context.Context, id string) (*kv_config.KvConfig, error) {
	config, ok := m.configs[id]
	if !ok {
		return nil, appErrors.ErrNotFound
	}
	return config, nil
}

func (m *mockKvConfigRepo) FindByKey(_ context.Context, key string) (*kv_config.KvConfig, error) {
	for _, config := range m.configs {
		if config.Key == key {
			return config, nil
		}
	}
	return nil, nil
}

func (m *mockKvConfigRepo) Delete(_ context.Context, id string) error {
	if _, ok := m.configs[id]; !ok {
		return appErrors.ErrNotFound
	}
	delete(m.configs, id)
	return nil
}

// ---------- tests ----------

func newTestKvConfigHandler() (*KvConfigCommandHandler, *mockKvConfigRepo) {
	repo := newMockKvConfigRepo()
	handler := NewKvConfigCommandHandler(repo)
	return handler, repo
}

func TestKvConfigCommandHandler_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		handler, _ := newTestKvConfigHandler()

		result, err := handler.Create(context.Background(), &command.CreateKvConfigCommand{
			Key: "site_name", Value: "TestApp", Description: "站点名称",
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result.Key != "site_name" {
			t.Errorf("expected key 'site_name', got %q", result.Key)
		}
		if result.Value != "TestApp" {
			t.Errorf("expected value 'TestApp', got %q", result.Value)
		}
	})

	t.Run("duplicate key", func(t *testing.T) {
		handler, repo := newTestKvConfigHandler()
		kv, _ := kv_config.NewKvConfig("existing", "site_name", "Old", "")
		repo.configs["existing"] = kv

		_, err := handler.Create(context.Background(), &command.CreateKvConfigCommand{
			Key: "site_name", Value: "New",
		})
		if err != ErrKvConfigKeyDuplicate {
			t.Errorf("expected ErrKvConfigKeyDuplicate, got %v", err)
		}
	})

	t.Run("empty key domain validation", func(t *testing.T) {
		handler, _ := newTestKvConfigHandler()
		_, err := handler.Create(context.Background(), &command.CreateKvConfigCommand{
			Key: "", Value: "test",
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestKvConfigCommandHandler_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		handler, repo := newTestKvConfigHandler()
		kv, _ := kv_config.NewKvConfig("id-1", "site_name", "TestApp", "")
		repo.configs["id-1"] = kv

		result, err := handler.Update(context.Background(), &command.UpdateKvConfigCommand{
			ID: "id-1", Value: "NewName", Description: "新描述",
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result.Value != "NewName" {
			t.Errorf("expected value 'NewName', got %q", result.Value)
		}
		if result.Description != "新描述" {
			t.Errorf("expected description '新描述', got %q", result.Description)
		}
	})

	t.Run("not found", func(t *testing.T) {
		handler, _ := newTestKvConfigHandler()
		_, err := handler.Update(context.Background(), &command.UpdateKvConfigCommand{
			ID: "nonexistent", Value: "test",
		})
		if err != appErrors.ErrNotFound {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestKvConfigCommandHandler_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		handler, repo := newTestKvConfigHandler()
		kv, _ := kv_config.NewKvConfig("id-1", "site_name", "TestApp", "")
		repo.configs["id-1"] = kv

		err := handler.Delete(context.Background(), &command.DeleteKvConfigCommand{ID: "id-1"})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if _, ok := repo.configs["id-1"]; ok {
			t.Error("expected config to be deleted from repo")
		}
	})

	t.Run("not found", func(t *testing.T) {
		handler, _ := newTestKvConfigHandler()
		err := handler.Delete(context.Background(), &command.DeleteKvConfigCommand{ID: "nonexistent"})
		if err != appErrors.ErrNotFound {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}