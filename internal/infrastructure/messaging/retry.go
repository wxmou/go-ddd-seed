package messaging

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"

	appEvent "github.com/go-ddd-seed/go-ddd-seed/internal/application/event"
)

// PoisonQueueConfig 死信队列配置
type PoisonQueueConfig struct {
	// Topic 死信队列 topic 名称
	Topic string
	// Publisher 用于发送死信消息的 Publisher
	Publisher message.Publisher
}

// MQMiddleware MQ 中间件工厂
type MQMiddleware struct {
	logger watermill.LoggerAdapter
}

// NewMQMiddleware 创建 MQ 中间件工厂
func NewMQMiddleware(logger watermill.LoggerAdapter) *MQMiddleware {
	return &MQMiddleware{logger: logger}
}

// WithRetry 创建重试中间件
// 使用指数退避策略，重试耗尽后返回错误
func (m *MQMiddleware) WithRetry(config appEvent.RetryConfig) message.HandlerMiddleware {
	return middleware.Retry{
		MaxRetries:      config.MaxRetries,
		InitialInterval: config.InitialBackoff,
		MaxInterval:     config.MaxBackoff,
		Multiplier:      2.0,
		Logger:          m.logger,
	}.Middleware
}

// WithPoisonQueue 创建死信队列中间件
// 当消息重试耗尽后，将其发送到死信 topic
func (m *MQMiddleware) WithPoisonQueue(config PoisonQueueConfig) message.HandlerMiddleware {
	poison, err := middleware.PoisonQueue(config.Publisher, config.Topic)
	if err != nil {
		// PoisonQueue 在创建时如果配置无效会返回 error
		// 这里使用一个兜底：返回一个空中间件（直接透传）
		return func(h message.HandlerFunc) message.HandlerFunc {
			return h
		}
	}
	return poison
}

// WithInstantAck 创建即时确认中间件
// 收到消息后立即 ACK，避免消息处理超时导致重入
func (m *MQMiddleware) WithInstantAck() message.HandlerMiddleware {
	return middleware.InstantAck
}

// WithCorrelationID 创建关联 ID 中间件
// 自动为消息添加关联 ID，方便追踪
func (m *MQMiddleware) WithCorrelationID() message.HandlerMiddleware {
	return middleware.CorrelationID
}