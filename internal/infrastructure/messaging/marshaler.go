// Package mq 消息队列基础设施实现
// 基于 Watermill 的适配器，可接入任意 MQ 中间件
package messaging

import (
	"encoding/json"

	domainEvent "github.com/go-ddd-seed/go-ddd-seed/internal/domain/event"
)

// DomainEventMarshaler 领域事件序列化器接口
// 将 DomainEvent 序列化为 []byte 用于 MQ 传输
type DomainEventMarshaler interface {
	// Marshal 将领域事件序列化为字节
	Marshal(event domainEvent.DomainEvent) ([]byte, error)
}

// JSONMarshaler JSON 序列化器
type JSONMarshaler struct{}

// NewJSONMarshaler 创建 JSON 序列化器
func NewJSONMarshaler() *JSONMarshaler {
	return &JSONMarshaler{}
}

// Marshal 将领域事件序列化为 JSON 字节
func (m *JSONMarshaler) Marshal(event domainEvent.DomainEvent) ([]byte, error) {
	return json.Marshal(event)
}