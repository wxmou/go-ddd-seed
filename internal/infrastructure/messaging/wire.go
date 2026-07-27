package messaging

import (
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/google/wire"

	appEvent "github.com/go-ddd-seed/go-ddd-seed/internal/application/event"
)

// MQSet Wire 依赖集（开发/测试环境）
// 使用 Go Channel Pub/Sub，零外部依赖
var MQSet = wire.NewSet(
	// 日志适配器
	NewWatermillLogger,

	// Go Channel（开发/测试环境）
	NewGoChannelPubSub,
	wire.Bind(new(message.Publisher), new(*gochannel.GoChannel)),
	wire.Bind(new(message.Subscriber), new(*gochannel.GoChannel)),

	// 序列化器
	NewJSONMarshaler,
	wire.Bind(new(DomainEventMarshaler), new(*JSONMarshaler)),

	// 事件总线
	NewWatermillEventBus,
	wire.Bind(new(appEvent.EventBus), new(*WatermillEventBus)),

	// 中间件工厂
	NewMQMiddleware,
)

// RedisStreamMQSet Wire 依赖集（生产环境）
// 使用 Redis Stream Pub/Sub，需提前配置 Redis 连接
// 注意：NewRedisStreamSubscriber 需要额外配置参数，在 wire.go 中手动提供
var RedisStreamMQSet = wire.NewSet(
	// 日志适配器
	NewWatermillLogger,

	// Redis Stream Publisher
	NewRedisStreamPublisher,
	wire.Bind(new(message.Publisher), new(*redisstream.Publisher)),

	// 序列化器
	NewJSONMarshaler,
	wire.Bind(new(DomainEventMarshaler), new(*JSONMarshaler)),

	// 事件总线
	NewWatermillEventBus,
	wire.Bind(new(appEvent.EventBus), new(*WatermillEventBus)),

	// 中间件工厂
	NewMQMiddleware,
)