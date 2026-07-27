package kv_config

import (
	"time"

	"github.com/go-ddd-seed/go-ddd-seed/internal/domain"
	"github.com/go-ddd-seed/go-ddd-seed/internal/domain/event"
	"github.com/go-ddd-seed/go-ddd-seed/internal/domain/model"
)

// KvConfigStatus 键值配置状态
const (
	KvConfigStatusDisabled = 0
	KvConfigStatusEnabled  = 1
)

// Event names
const (
	EventKvConfigCreated = "kv.config.created"
	EventKvConfigUpdated = "kv.config.updated"
	EventKvConfigDeleted = "kv.config.deleted"
)

// KvConfigUpdatedEvent KV 配置更新事件
type KvConfigUpdatedEvent struct {
	event.BaseEvent
	AggregateID string
	Key         string
	OldValue    string
	NewValue    string
}

// NewKvConfigUpdatedEvent 创建 KV 配置更新事件
func NewKvConfigUpdatedEvent(aggregateID, key, oldValue, newValue string) KvConfigUpdatedEvent {
	return KvConfigUpdatedEvent{
		BaseEvent: event.BaseEvent{
			At:   time.Now(),
			Name: EventKvConfigUpdated,
		},
		AggregateID: aggregateID,
		Key:         key,
		OldValue:    oldValue,
		NewValue:    newValue,
	}
}

// KvConfig 键值配置聚合根
type KvConfig struct {
	model.AggregateRoot
	ID          string
	Key         string
	Value       string
	Description string
	Status      int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewKvConfig 创建新的键值配置
func NewKvConfig(id, key, value, description string) (*KvConfig, error) {
	if key == "" {
		return nil, ErrKvConfigKeyEmpty
	}
	now := time.Now()
	return &KvConfig{
		ID:          id,
		Key:         key,
		Value:       value,
		Description: description,
		Status:      KvConfigStatusEnabled,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// Update 更新配置值和描述
func (k *KvConfig) Update(value, description string) {
	oldValue := k.Value
	k.Value = value
	k.Description = description
	k.UpdatedAt = time.Now()

	k.AddDomainEvent(NewKvConfigUpdatedEvent(k.ID, k.Key, oldValue, value))
}

// Enable 启用配置
func (k *KvConfig) Enable() error {
	if k.Status == KvConfigStatusEnabled {
		return ErrKvConfigAlreadyEnabled
	}
	k.Status = KvConfigStatusEnabled
	k.UpdatedAt = time.Now()
	return nil
}

// Disable 禁用配置
func (k *KvConfig) Disable() error {
	if k.Status == KvConfigStatusDisabled {
		return ErrKvConfigAlreadyDisabled
	}
	k.Status = KvConfigStatusDisabled
	k.UpdatedAt = time.Now()
	return nil
}

// IsEnabled 是否启用
func (k *KvConfig) IsEnabled() bool {
	return k.Status == KvConfigStatusEnabled
}

// KvConfig 聚合错误码范围: 2100-2199
var (
	ErrKvConfigKeyEmpty        = domain.NewDomainError(2100, "键值配置 key 不能为空")
	ErrKvConfigAlreadyEnabled  = domain.NewDomainError(2101, "键值配置已启用")
	ErrKvConfigAlreadyDisabled = domain.NewDomainError(2102, "键值配置已禁用")
)