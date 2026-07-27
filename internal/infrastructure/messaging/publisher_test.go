package messaging

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"

	domainEvent "github.com/go-ddd-seed/go-ddd-seed/internal/domain/event"
)

// mockPublisher 模拟 Watermill Publisher
type mockPublisher struct {
	published   message.Messages
	publishFunc func(topic string, msgs ...*message.Message) error
}

func (m *mockPublisher) Publish(topic string, msgs ...*message.Message) error {
	m.published = append(m.published, msgs...)
	if m.publishFunc != nil {
		return m.publishFunc(topic, msgs...)
	}
	return nil
}

func (m *mockPublisher) Close() error { return nil }

// failingMarshaler 故意失败的序列化器
type failingMarshaler struct{}

func (f *failingMarshaler) Marshal(event domainEvent.DomainEvent) ([]byte, error) {
	return nil, errors.New("marshal failed")
}

func TestWatermillEventBus_Publish(t *testing.T) {
	logger := watermill.NopLogger{}

	t.Run("publish single event successfully", func(t *testing.T) {
		pub := &mockPublisher{}
		marshaler := NewJSONMarshaler()
		bus := NewWatermillEventBus(pub, marshaler, logger)

		event := newTestEvent("kv.config.updated", "id-1", "site_name", "TestApp")
		err := bus.Publish(context.Background(), event)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(pub.published) != 1 {
			t.Fatalf("expected 1 published message, got %d", len(pub.published))
		}

		msg := pub.published[0]
		if msg.Metadata.Get("event_name") != "kv.config.updated" {
			t.Errorf("expected metadata event_name 'kv.config.updated', got %q",
				msg.Metadata.Get("event_name"))
		}
		if msg.UUID == "" {
			t.Error("expected non-empty message UUID")
		}
	})

	t.Run("publish multiple events", func(t *testing.T) {
		pub := &mockPublisher{}
		marshaler := NewJSONMarshaler()
		bus := NewWatermillEventBus(pub, marshaler, logger)

		events := []domainEvent.DomainEvent{
			newTestEvent("kv.config.updated", "id-1", "k1", "v1"),
			newTestEvent("kv.config.updated", "id-2", "k2", "v2"),
			newTestEvent("user.registered", "u-1", "email", "test@test.com"),
		}

		err := bus.Publish(context.Background(), events...)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(pub.published) != 3 {
			t.Fatalf("expected 3 published messages, got %d", len(pub.published))
		}

		// 验证每个消息的 metadata
		expectedNames := []string{"kv.config.updated", "kv.config.updated", "user.registered"}
		for i, msg := range pub.published {
			if msg.Metadata.Get("event_name") != expectedNames[i] {
				t.Errorf("message[%d]: expected event_name %q, got %q",
					i, expectedNames[i], msg.Metadata.Get("event_name"))
			}
		}
	})

	t.Run("publish with marshal error", func(t *testing.T) {
		pub := &mockPublisher{}
		bus := NewWatermillEventBus(pub, &failingMarshaler{}, logger)

		event := newTestEvent("test.event", "id-1", "k", "v")
		err := bus.Publish(context.Background(), event)
		if err == nil {
			t.Fatal("expected error from marshal failure, got nil")
		}
	})

	t.Run("publish with publisher error", func(t *testing.T) {
		pub := &mockPublisher{
			publishFunc: func(topic string, msgs ...*message.Message) error {
				return errors.New("publish failed")
			},
		}
		marshaler := NewJSONMarshaler()
		bus := NewWatermillEventBus(pub, marshaler, logger)

		event := newTestEvent("test.event", "id-1", "k", "v")
		err := bus.Publish(context.Background(), event)
		if err == nil {
			t.Fatal("expected error from publisher failure, got nil")
		}
	})

	t.Run("publish empty events", func(t *testing.T) {
		pub := &mockPublisher{}
		marshaler := NewJSONMarshaler()
		bus := NewWatermillEventBus(pub, marshaler, logger)

		err := bus.Publish(context.Background())
		if err != nil {
			t.Fatalf("expected no error for empty events, got %v", err)
		}
		if len(pub.published) != 0 {
			t.Errorf("expected 0 published messages, got %d", len(pub.published))
		}
	})

	t.Run("publish uses EventName as topic", func(t *testing.T) {
		pub := &mockPublisher{}
		marshaler := NewJSONMarshaler()
		bus := NewWatermillEventBus(pub, marshaler, logger)

		// 使用实际存在的领域事件名（来自 kv_config 包）
		event := &testEvent{
			BaseEvent: domainEvent.BaseEvent{
				At:   time.Now(),
				Name: "kv.config.updated",
			},
			AggregateID: "test-id",
			Key:         "test-key",
			Value:       "test-value",
		}

		err := bus.Publish(context.Background(), event)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// 验证发布到正确的 topic
		if len(pub.published) != 1 {
			t.Fatal("expected 1 message")
		}
	})
}

func TestWatermillEventBus_ImplementsEventBus(t *testing.T) {
	// 编译期检查：WatermillEventBus 实现了 EventBus 接口
	var _ interface {
		Publish(ctx context.Context, events ...domainEvent.DomainEvent) error
	} = (*WatermillEventBus)(nil)
}

func TestNewWatermillEventBus(t *testing.T) {
	pub := &mockPublisher{}
	marshaler := NewJSONMarshaler()
	logger := watermill.NopLogger{}

	bus := NewWatermillEventBus(pub, marshaler, logger)
	if bus == nil {
		t.Fatal("expected non-nil WatermillEventBus")
	}
	if bus.publisher != pub {
		t.Error("expected publisher to be set")
	}
	if bus.marshaler != marshaler {
		t.Error("expected marshaler to be set")
	}
}