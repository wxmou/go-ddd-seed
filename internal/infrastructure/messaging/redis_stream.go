package messaging

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/redis/go-redis/v9"
)

// RedisStreamConfig Redis Stream 配置
type RedisStreamConfig struct {
	// ConsumerGroup 消费者组名称
	ConsumerGroup string
	// Consumer 消费者 ID（同一组内唯一），默认为 hostname
	Consumer string
}

// NewRedisStreamPublisher 创建 Redis Stream Publisher（生产-轻量级）
func NewRedisStreamPublisher(rdb *redis.Client, logger watermill.LoggerAdapter) (*redisstream.Publisher, error) {
	config := redisstream.PublisherConfig{
		Client: rdb,
	}
	return redisstream.NewPublisher(config, logger)
}

// NewRedisStreamSubscriber 创建 Redis Stream Subscriber（生产-轻量级）
func NewRedisStreamSubscriber(
	rdb *redis.Client,
	cfg RedisStreamConfig,
	logger watermill.LoggerAdapter,
) (*redisstream.Subscriber, error) {
	consumer := cfg.Consumer
	if consumer == "" {
		consumer = "app-consumer"
	}
	consumerGroup := cfg.ConsumerGroup
	if consumerGroup == "" {
		consumerGroup = "app"
	}

	config := redisstream.SubscriberConfig{
		Client:        rdb,
		Consumer:      consumer,
		ConsumerGroup: consumerGroup,
	}
	return redisstream.NewSubscriber(config, logger)
}
