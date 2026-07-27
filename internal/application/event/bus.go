// Package event 应用层事件总线端口
// 只负责事件发布，不耦合订阅语义
package event

import (
	"context"
	"time"

	domainEvent "github.com/go-ddd-seed/go-ddd-seed/internal/domain/event"
)

// EventBus 事件总线接口
// 只负责发布事件，监听由 trigger/listener/ 层处理
type EventBus interface {
	Publish(ctx context.Context, events ...domainEvent.DomainEvent) error
}

// RetryConfig 重试策略配置
// 定义在应用层端口，基础设施层负责具体实现
type RetryConfig struct {
	// MaxRetries 最大重试次数
	MaxRetries int
	// InitialBackoff 首次重试间隔
	InitialBackoff time.Duration
	// MaxBackoff 最大重试间隔
	MaxBackoff time.Duration
}

// DefaultRetryConfig 默认重试配置
var DefaultRetryConfig = RetryConfig{
	MaxRetries:     3,
	InitialBackoff: 100 * time.Millisecond,
	MaxBackoff:     10 * time.Second,
}

// RouteHandler 消息路由处理器
// 接收原始消息字节，转发给 Listener 的 HandleMessage
type RouteHandler func(ctx context.Context, raw []byte) error

// RouterConfig 路由配置项
// 定义 topic → handler 的映射关系，以及重试/死信策略
type RouterConfig struct {
	// RetryConfig 重试策略，零值使用默认
	RetryConfig RetryConfig
	// PoisonTopic 死信队列 topic，空则不启用
	PoisonTopic string
	// Handlers topic -> Handler 映射
	Handlers map[string]RouteHandler
}