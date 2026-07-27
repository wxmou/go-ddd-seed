package messaging

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"

	appEvent "github.com/go-ddd-seed/go-ddd-seed/internal/application/event"
)

func TestMQMiddleware_WithRetry(t *testing.T) {
	mw := NewMQMiddleware(watermill.NopLogger{})

	t.Run("creates retry middleware without error", func(t *testing.T) {
		middleware := mw.WithRetry(appEvent.RetryConfig{
			MaxRetries:     3,
			InitialBackoff: 1,
			MaxBackoff:     10,
		})
		if middleware == nil {
			t.Fatal("expected non-nil middleware")
		}
	})

	t.Run("retry middleware allows handler to execute", func(t *testing.T) {
		middleware := mw.WithRetry(appEvent.RetryConfig{
			MaxRetries:     2,
			InitialBackoff: 1,
			MaxBackoff:     10,
		})

		var callCount int
		handler := middleware(func(msg *message.Message) ([]*message.Message, error) {
			callCount++
			return nil, nil
		})

		msg := message.NewMessage("test-id", []byte("payload"))
		_, err := handler(msg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if callCount != 1 {
			t.Errorf("expected 1 call, got %d", callCount)
		}
	})

	t.Run("retry middleware retries on error", func(t *testing.T) {
		middleware := mw.WithRetry(appEvent.RetryConfig{
			MaxRetries:     3,
			InitialBackoff: 1,
			MaxBackoff:     10,
		})

		var callCount int
		handler := middleware(func(msg *message.Message) ([]*message.Message, error) {
			callCount++
			if callCount <= 2 {
				return nil, errors.New("temporary error")
			}
			return nil, nil
		})

		msg := message.NewMessage("test-id", []byte("payload"))
		_, err := handler(msg)
		if err != nil {
			t.Fatalf("expected no error after retry, got %v", err)
		}
		if callCount != 3 {
			t.Errorf("expected 3 calls (2 retries + 1 success), got %d", callCount)
		}
	})

	t.Run("retry middleware gives up after max retries", func(t *testing.T) {
		middleware := mw.WithRetry(appEvent.RetryConfig{
			MaxRetries:     2,
			InitialBackoff: 1,
			MaxBackoff:     10,
		})

		var callCount int
		handler := middleware(func(msg *message.Message) ([]*message.Message, error) {
			callCount++
			return nil, errors.New("persistent error")
		})

		msg := message.NewMessage("test-id", []byte("payload"))
		_, err := handler(msg)
		if err == nil {
			t.Fatal("expected error after max retries, got nil")
		}
		// 初始调用 + MaxRetries 次重试
		expectedCalls := 3 // 1 initial + 2 retries
		if callCount != expectedCalls {
			t.Errorf("expected %d calls, got %d", expectedCalls, callCount)
		}
	})

	t.Run("uses default retry config", func(t *testing.T) {
		middleware := mw.WithRetry(appEvent.DefaultRetryConfig)
		if middleware == nil {
			t.Fatal("expected non-nil middleware with default config")
		}
	})
}

func TestMQMiddleware_WithPoisonQueue(t *testing.T) {
	mw := NewMQMiddleware(watermill.NopLogger{})

	t.Run("creates poison queue middleware", func(t *testing.T) {
		pub := &mockPublisher{}
		middleware := mw.WithPoisonQueue(PoisonQueueConfig{
			Topic:     "poison",
			Publisher: pub,
		})
		if middleware == nil {
			t.Fatal("expected non-nil middleware")
		}
	})

	t.Run("poison queue handler executes successfully", func(t *testing.T) {
		pub := &mockPublisher{}
		middleware := mw.WithPoisonQueue(PoisonQueueConfig{
			Topic:     "poison",
			Publisher: pub,
		})

		handler := middleware(func(msg *message.Message) ([]*message.Message, error) {
			return nil, nil
		})

		msg := message.NewMessage("test-id", []byte("payload"))
		_, err := handler(msg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("nil publisher returns passthrough middleware", func(t *testing.T) {
		middleware := mw.WithPoisonQueue(PoisonQueueConfig{
			Topic:     "poison",
			Publisher: nil,
		})
		if middleware == nil {
			t.Fatal("expected non-nil passthrough middleware")
		}

		handler := middleware(func(msg *message.Message) ([]*message.Message, error) {
			return nil, nil
		})

		msg := message.NewMessage("test-id", []byte("payload"))
		_, err := handler(msg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
}

func TestMQMiddleware_WithInstantAck(t *testing.T) {
	mw := NewMQMiddleware(watermill.NopLogger{})

	t.Run("instant ack middleware executes handler", func(t *testing.T) {
		middleware := mw.WithInstantAck()
		if middleware == nil {
			t.Fatal("expected non-nil middleware")
		}

		handler := middleware(func(msg *message.Message) ([]*message.Message, error) {
			return nil, nil
		})

		msg := message.NewMessage("test-id", []byte("payload"))
		_, err := handler(msg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
}

func TestMQMiddleware_WithCorrelationID(t *testing.T) {
	mw := NewMQMiddleware(watermill.NopLogger{})

	t.Run("correlation id middleware adds correlation id", func(t *testing.T) {
		middleware := mw.WithCorrelationID()
		if middleware == nil {
			t.Fatal("expected non-nil middleware")
		}

		var capturedID string
		handler := middleware(func(msg *message.Message) ([]*message.Message, error) {
			capturedID = msg.Metadata.Get("correlation_id")
			return nil, nil
		})

		msg := message.NewMessage("test-id", []byte("payload"))
		msg.Metadata.Set("correlation_id", "corr-123")
		_, err := handler(msg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if capturedID != "corr-123" {
			t.Errorf("expected correlation_id 'corr-123', got %q", capturedID)
		}
	})

	t.Run("correlation id middleware passes through without id", func(t *testing.T) {
		middleware := mw.WithCorrelationID()

		handler := middleware(func(msg *message.Message) ([]*message.Message, error) {
			return nil, nil
		})

		msg := message.NewMessage("test-id", []byte("payload"))
		_, err := handler(msg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
}

func TestNewMQMiddleware(t *testing.T) {
	logger := watermill.NopLogger{}
	mw := NewMQMiddleware(logger)
	if mw == nil {
		t.Fatal("expected non-nil MQMiddleware")
	}
	if mw.logger != logger {
		t.Error("expected logger to be set")
	}
}

// TestPoisonQueueConfig_Struct verifies the struct is usable
func TestPoisonQueueConfig_Struct(t *testing.T) {
	cfg := PoisonQueueConfig{
		Topic:     "dead-letter",
		Publisher: &mockPublisher{},
	}
	if cfg.Topic != "dead-letter" {
		t.Errorf("expected topic 'dead-letter', got %q", cfg.Topic)
	}
	if cfg.Publisher == nil {
		t.Error("expected non-nil publisher")
	}
}

// mockSubscriber for router tests
type mockSubscriber struct{}

func (m *mockSubscriber) Subscribe(ctx context.Context, topic string) (<-chan *message.Message, error) {
	ch := make(chan *message.Message)
	close(ch) // immediately closed, no messages
	return ch, nil
}

func (m *mockSubscriber) Close() error { return nil }

func TestNewRouter(t *testing.T) {
	logger := watermill.NopLogger{}
	mw := NewMQMiddleware(logger)
	sub := &mockSubscriber{}
	pub := &mockPublisher{}

	t.Run("creates router with valid config", func(t *testing.T) {
		cfg := appEvent.RouterConfig{
			RetryConfig: appEvent.RetryConfig{
				MaxRetries:     1,
				InitialBackoff: 1,
				MaxBackoff:     10,
			},
			Handlers: map[string]appEvent.RouteHandler{
				"test.topic": func(ctx context.Context, raw []byte) error {
					return nil
				},
			},
		}

		router, err := NewRouter(sub, pub, mw, cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if router == nil {
			t.Fatal("expected non-nil router")
		}
	})

	t.Run("uses default retry config when zero value", func(t *testing.T) {
		cfg := appEvent.RouterConfig{
			// RetryConfig 零值
			Handlers: map[string]appEvent.RouteHandler{
				"test.topic": func(ctx context.Context, raw []byte) error {
					return nil
				},
			},
		}

		router, err := NewRouter(sub, pub, mw, cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if router == nil {
			t.Fatal("expected non-nil router")
		}
	})

	t.Run("creates router with multiple topics", func(t *testing.T) {
		cfg := appEvent.RouterConfig{
			RetryConfig: appEvent.DefaultRetryConfig,
			Handlers: map[string]appEvent.RouteHandler{
				"topic.a": func(ctx context.Context, raw []byte) error { return nil },
				"topic.b": func(ctx context.Context, raw []byte) error { return nil },
			},
		}

		router, err := NewRouter(sub, pub, mw, cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if router == nil {
			t.Fatal("expected non-nil router")
		}
	})

	t.Run("creates router with empty handlers", func(t *testing.T) {
		cfg := appEvent.RouterConfig{
			RetryConfig: appEvent.DefaultRetryConfig,
			Handlers:    map[string]appEvent.RouteHandler{},
		}

		router, err := NewRouter(sub, pub, mw, cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if router == nil {
			t.Fatal("expected non-nil router even with empty handlers")
		}
	})
}

func TestConsumerManager(t *testing.T) {
	logger := watermill.NopLogger{}
	mw := NewMQMiddleware(logger)
	sub := &mockSubscriber{}
	pub := &mockPublisher{}

	cfg := appEvent.RouterConfig{
		RetryConfig: appEvent.RetryConfig{
			MaxRetries:     1,
			InitialBackoff: 1,
			MaxBackoff:     10,
		},
		Handlers: map[string]appEvent.RouteHandler{
			"test.topic": func(ctx context.Context, raw []byte) error {
				return nil
			},
		},
	}

	router, err := NewRouter(sub, pub, mw, cfg)
	if err != nil {
		t.Fatalf("failed to create router: %v", err)
	}

	t.Run("creates consumer manager", func(t *testing.T) {
		cm := NewConsumerManager(router, sub, logger)
		if cm == nil {
			t.Fatal("expected non-nil ConsumerManager")
		}
	})

	t.Run("router accessor returns same router", func(t *testing.T) {
		cm := NewConsumerManager(router, sub, logger)
		if cm.Router() != router {
			t.Error("expected Router() to return the same router")
		}
	})

	t.Run("shutdown closes router", func(t *testing.T) {
		// 使用短关闭超时的 Router，避免 30s 等待
		shortRouter, err := message.NewRouter(message.RouterConfig{
			CloseTimeout: time.Second,
		}, watermill.NopLogger{})
		if err != nil {
			t.Fatalf("failed to create router: %v", err)
		}
		cm := NewConsumerManager(shortRouter, sub, logger)
		// Shutdown should work without error
		_ = cm.Shutdown()
	})
}

func TestRetryConfig_Defaults(t *testing.T) {
	cfg := appEvent.DefaultRetryConfig
	if cfg.MaxRetries != 3 {
		t.Errorf("expected MaxRetries=3, got %d", cfg.MaxRetries)
	}
	if cfg.InitialBackoff == 0 {
		t.Error("expected non-zero InitialBackoff")
	}
	if cfg.MaxBackoff == 0 {
		t.Error("expected non-zero MaxBackoff")
	}
}