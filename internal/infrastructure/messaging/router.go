package messaging

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"

	appEvent "github.com/go-ddd-seed/go-ddd-seed/internal/application/event"
)

// NewRouter 创建 Watermill Router 并配置路由
// 为每个 topic 注册 handler，添加重试/死信中间件
// 属于基础设施层，封装 Watermill 的创建细节
func NewRouter(
	subscriber message.Subscriber,
	publisher message.Publisher,
	middlewareFactory *MQMiddleware,
	cfg appEvent.RouterConfig,
) (*message.Router, error) {
	logger := watermill.NewStdLogger(false, false)

	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		return nil, err
	}

	// 添加重试中间件
	retryConfig := cfg.RetryConfig
	if retryConfig.MaxRetries == 0 {
		retryConfig = appEvent.DefaultRetryConfig
	}
	router.AddMiddleware(middlewareFactory.WithRetry(retryConfig))

	// 添加关联 ID 中间件
	router.AddMiddleware(middlewareFactory.WithCorrelationID())

	// 注册每个 topic 的 handler
	for topic, handler := range cfg.Handlers {
		h := handler // 捕获
		router.AddNoPublisherHandler(
			"handler-"+topic,
			topic,
			subscriber,
			func(msg *message.Message) error {
				return h(msg.Context(), msg.Payload)
			},
		)
		logger.Info("消息路由已注册", watermill.LogFields{
			"topic": topic,
		})
	}

	return router, nil
}