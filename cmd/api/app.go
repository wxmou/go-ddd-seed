package main

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/go-ddd-seed/go-ddd-seed/internal/infrastructure/adapter/repo"
	"github.com/go-ddd-seed/go-ddd-seed/internal/infrastructure/persistence"
	"github.com/go-ddd-seed/go-ddd-seed/internal/trigger/http/controller"
	"github.com/go-ddd-seed/go-ddd-seed/pkg/config"
	"github.com/go-ddd-seed/go-ddd-seed/pkg/logger"
	"github.com/go-ddd-seed/go-ddd-seed/pkg/middleware"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"

	_ "github.com/go-ddd-seed/go-ddd-seed/docs/swagger"
)

// App 应用结构体
type App struct {
	Engine *gin.Engine
	Logger *logger.Logger
}

// NewApp 创建应用
func NewApp(
	cfg *config.Config,
	db *gorm.DB,
	rdb *redis.Client,
	healthCtrl *controller.HealthController,
	kvConfigCtrl *controller.KvConfigController,
	dictCtrl *controller.DictController,
	authCtrl *controller.AuthController,
	rbacCtrl *controller.RbacController,
	auditCtrl *controller.AuditController,
	fileCtrl *controller.FileController,
	auditRepo *repo.AuditLogRepository,
) *App {
	// 设置 Gin 模式
	gin.SetMode(cfg.Server.Mode)

	// 初始化结构化日志
	logCfg := logger.LogConfig{
		Level:         cfg.Log.Level,
		Format:        cfg.Log.Format,
		SlowThreshold: cfg.Log.SlowThreshold,
	}
	l := logger.New(logCfg)
	logger.SetDefault(l)

	engine := gin.New()

	// 全局中间件
	engine.Use(middleware.Trace())      // Trace ID 注入（最先执行）
	engine.Use(middleware.BodyCache())  // 请求体缓存（在 Logger 之前）
	engine.Use(middleware.Logger())     // 结构化日志（基于 zerolog）
	engine.Use(middleware.CORS())
	engine.Use(gin.Recovery())

	// Swagger API 文档
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 健康检查路由（不需要认证）
	engine.GET("/health", healthCtrl.Health)

	// 自动迁移数据库表
	if err := persistence.AutoMigrate(db); err != nil {
		l.Fatal("数据库迁移失败", err)
	}

	// 字典缓存预热
	if rdb != nil {
		if err := dictCtrl.WarmUp(context.Background()); err != nil {
			l.Warn("字典缓存预热失败", map[string]interface{}{"error": err.Error()})
		} else {
			l.Info("字典缓存预热完成")
		}
	}

	// API 路由组
	api := engine.Group("/api/v1")

	// 公开路由（无需认证）
	publicApi := engine.Group("/api/v1")
	authCtrl.RegisterRoutes(publicApi, api)

	// 需认证路由
	api.Use(middleware.Auth(cfg.JWT.Secret, rdb))
	api.Use(middleware.AuditLogger(auditRepo))

	// 注册业务路由
	kvConfigCtrl.RegisterRoutes(api)
	dictCtrl.RegisterRoutes(api)
	rbacCtrl.RegisterRoutes(api)
	auditCtrl.RegisterRoutes(api)
	fileCtrl.RegisterRoutes(api)

	return &App{Engine: engine, Logger: l}
}

// Run 启动 HTTP 服务
func (a *App) Run(addr string) error {
	return a.Engine.Run(addr)
}

// Cleanup 清理资源
func Cleanup() error {
	// 关闭数据库连接等
	return nil
}