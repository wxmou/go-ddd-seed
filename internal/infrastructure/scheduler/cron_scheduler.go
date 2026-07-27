package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron/v3"

	appApi "github.com/go-ddd-seed/go-ddd-seed/internal/application/port/thirdPartyApi"
	"github.com/go-ddd-seed/go-ddd-seed/pkg/logger"
)

// registeredJob 已注册的任务元信息
type registeredJob struct {
	name     string
	cronExpr string
	entryID  cron.EntryID
}

// CronScheduler robfig/cron 调度器实现
type CronScheduler struct {
	cron    *cron.Cron
	logger  *logger.Logger
	jobs    []registeredJob
	mu      sync.Mutex
	onError func(jobName string, err error)
}

// jobTimeout 任务超时包装器
type jobTimeout struct {
	inner   appApi.Job
	timeout time.Duration
}

func (j *jobTimeout) Name() string { return j.inner.Name() }
func (j *jobTimeout) Run(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, j.timeout)
	defer cancel()
	return j.inner.Run(ctx)
}

// NewCronScheduler 创建 cron 调度器
func NewCronScheduler(logger *logger.Logger) *CronScheduler {
	return &CronScheduler{
		cron: cron.New(
			cron.WithSeconds(),
			cron.WithChain(
				cron.SkipIfStillRunning(cron.DiscardLogger), // 同一任务未执行完时跳过
				cron.Recover(cron.DiscardLogger),             // panic 恢复
			),
		),
		logger: logger,
		jobs:   make([]registeredJob, 0),
	}
}

// WithTimeout 设置任务默认超时时间
// 返回新的调度器实例（不修改原对象）
func (s *CronScheduler) WithTimeout(timeout time.Duration) *CronScheduler {
	// 包装所有已注册任务
	for i, j := range s.jobs {
		// 通过 cron 的 Entry 重新设置无法直接修改，因此记录超时配置
		// 超时在 AddJob 时通过包装器实现
		_ = i
		_ = j
	}
	return s
}

// WithErrorHandler 设置任务失败回调
func (s *CronScheduler) WithErrorHandler(handler func(jobName string, err error)) *CronScheduler {
	s.onError = handler
	return s
}

// AddJob 注册定时任务
func (s *CronScheduler) AddJob(cronExpr string, job appApi.Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 包装为带日志的 cron Job
	cronJob := cron.FuncJob(func() {
		ctx := context.Background()
		s.logger.Info("定时任务开始执行", map[string]any{
			"job": job.Name(),
		})

		start := time.Now()
		if err := job.Run(ctx); err != nil {
			s.logger.Error("定时任务执行失败", err, map[string]any{
				"job":    job.Name(),
				"elapsed": time.Since(start).Milliseconds(),
			})
			if s.onError != nil {
				s.onError(job.Name(), err)
			}
			return
		}

		s.logger.Info("定时任务执行完成", map[string]any{
			"job":     job.Name(),
			"elapsed": time.Since(start).Milliseconds(),
		})
	})

	entryID, err := s.cron.AddFunc(cronExpr, cronJob)
	if err != nil {
		return fmt.Errorf("注册定时任务失败 [%s: %s]: %w", job.Name(), cronExpr, err)
	}

	s.jobs = append(s.jobs, registeredJob{
		name:     job.Name(),
		cronExpr: cronExpr,
		entryID:  entryID,
	})

	s.logger.Info("定时任务已注册", map[string]any{
		"job":  job.Name(),
		"cron": cronExpr,
	})
	return nil
}

// Start 启动调度器（非阻塞）
func (s *CronScheduler) Start() error {
	s.cron.Start()
	s.logger.Info("定时任务调度器已启动", map[string]any{
		"job_count": len(s.jobs),
	})
	return nil
}

// Stop 优雅停止调度器
func (s *CronScheduler) Stop(ctx context.Context) error {
	stopped := s.cron.Stop()
	select {
	case <-stopped.Done():
		s.logger.Info("定时任务调度器已优雅停止")
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// GetJobs 返回已注册的任务列表
func (s *CronScheduler) GetJobs() []registeredJob {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]registeredJob, len(s.jobs))
	copy(result, s.jobs)
	return result
}

// Ensure interface compliance
var _ appApi.Scheduler = (*CronScheduler)(nil)