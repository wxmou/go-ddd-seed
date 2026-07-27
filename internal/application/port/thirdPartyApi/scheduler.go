package thirdPartyApi

import "context"

// Job 定时任务接口
// 所有定时任务必须实现此接口
type Job interface {
	// Name 返回任务名称，用于日志和监控
	Name() string
	// Run 执行任务
	Run(ctx context.Context) error
}

// Scheduler 调度器接口
// 定义在应用层，基础设施层实现
// 当前使用 robfig/cron 单机实现，未来可替换为分布式调度
type Scheduler interface {
	// AddJob 注册定时任务
	// cronExpr: 标准 cron 表达式（5/6 字段，支持秒级）
	// job: 实现了 Job 接口的任务实例
	AddJob(cronExpr string, job Job) error

	// Start 启动调度器（非阻塞）
	Start() error

	// Stop 优雅停止调度器（等待正在执行的任务完成）
	Stop(ctx context.Context) error
}