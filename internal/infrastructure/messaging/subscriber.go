package messaging

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
)

// ConsumerManager 消费者管理器
// 管理 Watermill Router 的生命周期
type ConsumerManager struct {
	router     *message.Router
	subscriber message.Subscriber
	logger     watermill.LoggerAdapter
}

// NewConsumerManager 创建消费者管理器
func NewConsumerManager(
	router *message.Router,
	subscriber message.Subscriber,
	logger watermill.LoggerAdapter,
) *ConsumerManager {
	return &ConsumerManager{
		router:     router,
		subscriber: subscriber,
		logger:     logger,
	}
}

// Start 启动所有消费者（阻塞直到收到 SIGINT/SIGTERM）
func (m *ConsumerManager) Start(ctx context.Context) error {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		m.logger.Info("收到终止信号，正在优雅关闭消费者...", watermill.LogFields{
			"signal": sig.String(),
		})
		m.router.Close()
	}()

	m.logger.Info("消息消费者已启动", nil)
	return m.router.Run(ctx)
}

// Shutdown 优雅关闭（等待当前消息处理完成）
func (m *ConsumerManager) Shutdown() error {
	return m.router.Close()
}

// Router 返回内部 Router（用于注册 Handler 等操作）
func (m *ConsumerManager) Router() *message.Router {
	return m.router
}
