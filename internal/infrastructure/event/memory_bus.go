// Package event 事件总线基础设施实现
// 仅用于开发/测试环境
package event

import (
	"context"
	"fmt"
	"sync"

	appEvent "github.com/go-ddd-seed/go-ddd-seed/internal/application/event"
	domainEvent "github.com/go-ddd-seed/go-ddd-seed/internal/domain/event"
	"github.com/go-ddd-seed/go-ddd-seed/pkg/logger"
)

// EventHandler 事件处理函数（仅用于 InMemoryEventBus 注册）
// 生产环境下监听器通过 HandleMessage(raw []byte) 工作，不依赖此类型
type EventHandler func(ctx context.Context, event domainEvent.DomainEvent) error

// InMemoryEventBus 同步内存事件总线
// 仅用于开发/测试环境
type InMemoryEventBus struct {
	mu       sync.RWMutex
	handlers map[string][]EventHandler
}

// 编译期接口检查
var _ appEvent.EventBus = (*InMemoryEventBus)(nil)

// NewInMemoryEventBus 创建内存事件总线
func NewInMemoryEventBus() *InMemoryEventBus {
	return &InMemoryEventBus{
		handlers: make(map[string][]EventHandler),
	}
}

// Subscribe 订阅事件（辅助方法，不在 EventBus 接口中）
// 仅在开发/测试中使用
func (b *InMemoryEventBus) Subscribe(eventName string, handler EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventName] = append(b.handlers[eventName], handler)
}

// Publish 实现 EventBus 接口
func (b *InMemoryEventBus) Publish(ctx context.Context, events ...domainEvent.DomainEvent) error {
	for _, event := range events {
		if err := b.publishOne(ctx, event); err != nil {
			return err
		}
	}
	return nil
}

func (b *InMemoryEventBus) publishOne(ctx context.Context, event domainEvent.DomainEvent) error {
	b.mu.RLock()
	handlers := b.handlers[event.EventName()]
	b.mu.RUnlock()

	if len(handlers) == 0 {
		logger.Debug("领域事件无处理器", map[string]interface{}{
			"event": event.EventName(),
		})
		return nil
	}

	for _, handler := range handlers {
		if err := handler(ctx, event); err != nil {
			return fmt.Errorf("事件 %s 处理失败: %w", event.EventName(), err)
		}
	}
	return nil
}