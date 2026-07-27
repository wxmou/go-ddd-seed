// Package model 领域模型基类
// 提供领域事件记录能力
package model

import "github.com/go-ddd-seed/go-ddd-seed/internal/domain/event"

// AggregateRoot 聚合根基类
// 所有聚合根内嵌此结构体以支持领域事件
type AggregateRoot struct {
	domainEvents []event.DomainEvent
}

// AddDomainEvent 记录一个领域事件
func (a *AggregateRoot) AddDomainEvent(e event.DomainEvent) {
	a.domainEvents = append(a.domainEvents, e)
}

// ClearDomainEvents 清除并返回所有未发布的领域事件
func (a *AggregateRoot) ClearDomainEvents() []event.DomainEvent {
	events := a.domainEvents
	a.domainEvents = nil
	return events
}

// HasDomainEvents 是否有未发布的事件
func (a *AggregateRoot) HasDomainEvents() bool {
	return len(a.domainEvents) > 0
}

// DomainEventAware 表示聚合根能够产生领域事件
// 聚合根内嵌 AggregateRoot 后自动满足此接口
type DomainEventAware interface {
	ClearDomainEvents() []event.DomainEvent
	HasDomainEvents() bool
}