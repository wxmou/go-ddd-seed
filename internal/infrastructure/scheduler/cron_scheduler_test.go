package scheduler

import (
	"context"
	"sync"
	"testing"
	"time"

	appApi "github.com/go-ddd-seed/go-ddd-seed/internal/application/port/thirdPartyApi"
	"github.com/go-ddd-seed/go-ddd-seed/pkg/logger"
)

// testJob 用于测试的简单 Job
type testJob struct {
	name     string
	mu       sync.Mutex
	runCount int
	lastErr  error
	blockCh  chan struct{} // 用于模拟长时间运行
}

func newTestJob(name string) *testJob {
	return &testJob{name: name}
}

func (j *testJob) Name() string { return j.name }

func (j *testJob) Run(ctx context.Context) error {
	j.mu.Lock()
	j.runCount++
	j.mu.Unlock()

	if j.blockCh != nil {
		select {
		case <-j.blockCh:
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return j.lastErr
}

func (j *testJob) getRunCount() int {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.runCount
}

// testJobImplementsJob 编译期接口检查
var _ appApi.Job = (*testJob)(nil)

func newTestLogger() *logger.Logger {
	return logger.New(logger.LogConfig{
		Level:  "warn",
		Format: "text",
	})
}

func TestCronScheduler_AddJob(t *testing.T) {
	sched := NewCronScheduler(newTestLogger())

	job := newTestJob("test_job")
	err := sched.AddJob("* * * * * *", job) // 每秒执行
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	jobs := sched.GetJobs()
	if len(jobs) != 1 {
		t.Errorf("expected 1 registered job, got %d", len(jobs))
	}
	if jobs[0].name != "test_job" {
		t.Errorf("expected job name 'test_job', got %q", jobs[0].name)
	}
}

func TestCronScheduler_StartStop(t *testing.T) {
	sched := NewCronScheduler(newTestLogger())

	// 注册一个永远不会触发的任务
	job := newTestJob("test_job")
	if err := sched.AddJob("0 0 1 1 1 *", job); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := sched.Start(); err != nil {
		t.Fatalf("expected no error on start, got %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := sched.Stop(ctx); err != nil {
		t.Errorf("expected no error on stop, got %v", err)
	}
}

func TestCronScheduler_JobExecution(t *testing.T) {
	sched := NewCronScheduler(newTestLogger())

	job := newTestJob("exec_job")
	// 注册每秒执行一次的任务
	if err := sched.AddJob("* * * * * *", job); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := sched.Start(); err != nil {
		t.Fatalf("expected no error on start, got %v", err)
	}

	// 等待至少 2 秒让任务执行
	time.Sleep(2500 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := sched.Stop(ctx); err != nil {
		t.Errorf("expected no error on stop, got %v", err)
	}

	count := job.getRunCount()
	if count < 2 {
		t.Errorf("expected at least 2 executions, got %d", count)
	}
}

func TestCronScheduler_ErrorHandler(t *testing.T) {
	var capturedErr error
	var capturedJobName string

	sched := NewCronScheduler(newTestLogger()).
		WithErrorHandler(func(jobName string, err error) {
			capturedJobName = jobName
			capturedErr = err
		})

	job := newTestJob("error_job")
	job.lastErr = context.DeadlineExceeded

	if err := sched.AddJob("* * * * * *", job); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := sched.Start(); err != nil {
		t.Fatalf("expected no error on start, got %v", err)
	}

	// 等待任务执行
	time.Sleep(1500 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	sched.Stop(ctx)

	if capturedJobName != "error_job" {
		t.Errorf("expected captured job name 'error_job', got %q", capturedJobName)
	}
	if capturedErr != context.DeadlineExceeded {
		t.Errorf("expected captured error 'context deadline exceeded', got %v", capturedErr)
	}
}

func TestCronScheduler_InvalidCronExpr(t *testing.T) {
	sched := NewCronScheduler(newTestLogger())

	job := newTestJob("test_job")
	err := sched.AddJob("invalid-cron", job)
	if err == nil {
		t.Error("expected error for invalid cron expression, got nil")
	}
}

func TestCronScheduler_ImplementsInterface(t *testing.T) {
	// compile-time check
	var _ appApi.Scheduler = NewCronScheduler(newTestLogger())
}