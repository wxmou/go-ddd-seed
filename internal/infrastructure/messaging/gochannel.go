package messaging

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
)

// NewGoChannelPubSub 创建 Go Channel Pub/Sub 实现
// 用于开发/测试环境，零依赖
func NewGoChannelPubSub(logger watermill.LoggerAdapter) *gochannel.GoChannel {
	return gochannel.NewGoChannel(
		gochannel.Config{
			OutputChannelBuffer: 1024,
			Persistent:          false,
		},
		logger,
	)
}

// NewWatermillLogger 创建 Watermill 日志适配器
// 基于项目自定义 logger，但暂用标准输出
func NewWatermillLogger() watermill.LoggerAdapter {
	return watermill.NewStdLogger(false, false)
}