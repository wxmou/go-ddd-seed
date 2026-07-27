// Package event 领域事件契约
// 此包是项目最底层的叶子节点，只依赖标准库，禁止 import 任何业务聚合包
package event

import "time"

// DomainEvent 领域事件接口
// 所有领域事件必须实现此接口
// 具体事件类型定义在各聚合包内，内嵌 BaseEvent 实现此接口
type DomainEvent interface {
	EventName() string
	OccurredAt() time.Time
}

// BaseEvent 领域事件基类
// 具体事件结构体内嵌此类型以快速实现 DomainEvent 接口
type BaseEvent struct {
	At   time.Time
	Name string
}

// OccurredAt 返回事件发生时间
func (e BaseEvent) OccurredAt() time.Time { return e.At }

// EventName 返回事件名称
func (e BaseEvent) EventName() string { return e.Name }