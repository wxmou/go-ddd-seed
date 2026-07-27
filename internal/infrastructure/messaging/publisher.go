package messaging

import (
	"context"
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"

	appEvent "github.com/go-ddd-seed/go-ddd-seed/internal/application/event"
	domainEvent "github.com/go-ddd-seed/go-ddd-seed/internal/domain/event"
)

// WatermillEventBus 基于 Watermill 的事件总线实现
// 实现 application/event.EventBus 接口
type WatermillEventBus struct {
	publisher message.Publisher
	marshaler DomainEventMarshaler
	logger    watermill.LoggerAdapter
}

// 编译期检查接口实现
var _ appEvent.EventBus = (*WatermillEventBus)(nil)

// NewWatermillEventBus 创建 Watermill 事件总线
func NewWatermillEventBus(
	publisher message.Publisher,
	marshaler DomainEventMarshaler,
	logger watermill.LoggerAdapter,
) *WatermillEventBus {
	return &WatermillEventBus{
		publisher: publisher,
		marshaler: marshaler,
		logger:    logger,
	}
}

// Publish 实现 EventBus 接口
// 将 DomainEvent 序列化后通过 Watermill Publisher 发送
// topic 使用 DomainEvent.EventName() 作为路由键
func (b *WatermillEventBus) Publish(ctx context.Context, events ...domainEvent.DomainEvent) error {
	for _, event := range events {
		payload, err := b.marshaler.Marshal(event)
		if err != nil {
			return fmt.Errorf("序列化事件 %s 失败: %w", event.EventName(), err)
		}

		msg := message.NewMessage(watermill.NewUUID(), payload)
		msg.Metadata.Set("event_name", event.EventName())

		if err := b.publisher.Publish(event.EventName(), msg); err != nil {
			return fmt.Errorf("发布事件 %s 失败: %w", event.EventName(), err)
		}

		b.logger.Trace("事件已发布", watermill.LogFields{
			"event_name":  event.EventName(),
			"message_id":  msg.UUID,
		})
	}
	return nil
}