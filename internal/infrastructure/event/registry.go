package event

// HandlerRegistration 事件注册项
type HandlerRegistration struct {
	EventName string
	Handler   EventHandler
}

// EventRegistry 事件注册器（仅开发/测试环境）
// 将 EventHandler 注册到 InMemoryEventBus 上
type EventRegistry struct {
	bus *InMemoryEventBus
}

// NewEventRegistry 创建注册器并注册所有 Handler
func NewEventRegistry(bus *InMemoryEventBus, registrations ...HandlerRegistration) *EventRegistry {
	reg := &EventRegistry{bus: bus}
	for _, r := range registrations {
		bus.Subscribe(r.EventName, r.Handler)
	}
	return reg
}