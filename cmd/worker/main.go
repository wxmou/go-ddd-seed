package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-ddd-seed/go-ddd-seed/internal/infrastructure/adapter/repo"
	"github.com/go-ddd-seed/go-ddd-seed/internal/infrastructure/persistence"
	"github.com/go-ddd-seed/go-ddd-seed/internal/infrastructure/scheduler"
	"github.com/go-ddd-seed/go-ddd-seed/internal/trigger/job"
	"github.com/go-ddd-seed/go-ddd-seed/pkg/config"
	"github.com/go-ddd-seed/go-ddd-seed/pkg/logger"
	"gorm.io/gorm"
)

func main() {
	log.Println("Go DDD Seed Worker starting...")

	// 加载配置
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化日志
	logCfg := logger.LogConfig{
		Level:  cfg.Log.Level,
		Format: cfg.Log.Format,
	}
	appLogger := logger.New(logCfg)
	logger.SetDefault(appLogger)

	// 初始化数据库连接
	db, err := persistence.NewDatabase(&cfg.Database)
	if err != nil {
		appLogger.Fatal("数据库连接失败", err)
	}

	// 创建调度器
	sched := scheduler.NewCronScheduler(appLogger.WithModule("scheduler"))

	if cfg.Scheduler.Enabled {
		// 注册定时任务
		registerJobs(sched, db, cfg)
		// 启动调度器
		if err := sched.Start(); err != nil {
			appLogger.Fatal("调度器启动失败", err)
		}
	} else {
		appLogger.Info("定时任务调度器未启用")
	}

	// TODO: 注册消息队列消费者
	// mq := messaging.NewConsumer(...)
	// go mq.Start()

	// 等待退出信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("Go DDD Seed Worker shutting down...")

	// 优雅停止调度器
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := sched.Stop(ctx); err != nil {
		appLogger.Error("调度器停止失败", err)
	}
}

// registerJobs 注册所有定时任务
func registerJobs(sched *scheduler.CronScheduler, db *gorm.DB, cfg *config.Config) {
	// --- P1: 审计日志清理 ---
	auditRepo := repo.NewAuditLogRepository(db)
	auditCleanupJob := job.NewAuditCleanupJob(auditRepo, cfg.Scheduler.AuditRetentionDays)
	if err := sched.AddJob(cfg.Scheduler.AuditCleanupCron, auditCleanupJob); err != nil {
		// 注册失败记录日志但不阻塞启动
		logger.L().Error("审计日志清理任务注册失败", err)
	}

	// --- P2: 日报生成（待实现）---
	// dailyReportHandler := commandHandler.NewGenerateReportCommandHandler(...)
	// dailyReportJob := job.NewDailyReportJob(dailyReportHandler)
	// sched.AddJob("0 0 8 * * *", dailyReportJob)

	// --- P2: 月报统计（待实现）---
	// monthlyReportJob := job.NewMonthlyReportJob(...)
	// sched.AddJob("0 0 9 1 * *", monthlyReportJob)

	// --- P2: 缓存定时刷新（待实现）---
	// cacheWarmupJob := job.NewCacheWarmupJob(...)
	// sched.AddJob("0 0 */2 * * *", cacheWarmupJob)
}