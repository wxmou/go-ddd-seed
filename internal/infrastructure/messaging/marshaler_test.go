package messaging

import (
	"encoding/json"
	"testing"
	"time"

	domainEvent "github.com/go-ddd-seed/go-ddd-seed/internal/domain/event"
)

// testEvent 测试用领域事件
type testEvent struct {
	domainEvent.BaseEvent
	AggregateID string `json:"aggregate_id"`
	Key         string `json:"key"`
	Value       string `json:"value"`
}

func newTestEvent(name, aggregateID, key, value string) *testEvent {
	return &testEvent{
		BaseEvent: domainEvent.BaseEvent{
			At:   time.Now(),
			Name: name,
		},
		AggregateID: aggregateID,
		Key:         key,
		Value:       value,
	}
}

func TestJSONMarshaler_Marshal(t *testing.T) {
	m := NewJSONMarshaler()

	t.Run("marshal valid event", func(t *testing.T) {
		event := newTestEvent("kv.config.updated", "id-1", "site_name", "TestApp")
		raw, err := m.Marshal(event)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if raw == nil || len(raw) == 0 {
			t.Fatal("expected non-empty payload")
		}

		var decoded map[string]any
		if err := json.Unmarshal(raw, &decoded); err != nil {
			t.Fatalf("failed to unmarshal result: %v", err)
		}
		if decoded["aggregate_id"] != "id-1" {
			t.Errorf("expected aggregate_id 'id-1', got %v", decoded["aggregate_id"])
		}
		if decoded["key"] != "site_name" {
			t.Errorf("expected key 'site_name', got %v", decoded["key"])
		}
		if decoded["value"] != "TestApp" {
			t.Errorf("expected value 'TestApp', got %v", decoded["value"])
		}
	})

	t.Run("marshal event with embedded BaseEvent", func(t *testing.T) {
		event := newTestEvent("user.registered", "user-1", "email", "test@example.com")
		raw, err := m.Marshal(event)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		var decoded map[string]any
		json.Unmarshal(raw, &decoded)
		if decoded["Name"] != "user.registered" {
			t.Errorf("expected Name 'user.registered', got %v", decoded["Name"])
		}
	})

	t.Run("marshal event with empty fields", func(t *testing.T) {
		event := newTestEvent("", "", "", "")
		raw, err := m.Marshal(event)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if raw == nil {
			t.Fatal("expected non-nil payload even for empty event")
		}
	})
}

func TestDomainEventMarshaler_Interface(t *testing.T) {
	// 编译期检查：JSONMarshaler 实现了 DomainEventMarshaler 接口
	var _ DomainEventMarshaler = (*JSONMarshaler)(nil)

	m := NewJSONMarshaler()
	event := newTestEvent("test.event", "id-1", "k", "v")
	raw, err := m.Marshal(event)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(raw) == 0 {
		t.Fatal("expected non-empty marshal output")
	}
}