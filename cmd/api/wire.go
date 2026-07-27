//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	appRepo "github.com/go-ddd-seed/go-ddd-seed/internal/application/port/repo"
	"github.com/go-ddd-seed/go-ddd-seed/internal/application/commandHandler"
	"github.com/go-ddd-seed/go-ddd-seed/internal/application/queryService"
	domainRepo "github.com/go-ddd-seed/go-ddd-seed/internal/domain/port/repo"
	"github.com/go-ddd-seed/go-ddd-seed/internal/infrastructure/adapter/repo"
	storageAdapter "github.com/go-ddd-seed/go-ddd-seed/internal/infrastructure/adapter/thirdPartyApi/storage"
	"github.com/go-ddd-seed/go-ddd-seed/internal/infrastructure/messaging"
	"github.com/go-ddd-seed/go-ddd-seed/internal/infrastructure/persistence"
	"github.com/go-ddd-seed/go-ddd-seed/internal/trigger/http/controller"
	"github.com/go-ddd-seed/go-ddd-seed/pkg/config"
)

// InitializeApp 通过 Wire 注入依赖
func InitializeApp(cfg *config.Config) (*App, func(), error) {
	wire.Build(
		// 配置提取
		provideDatabaseConfig,
		provideRedisConfig,
		provideStorageConfig,
		provideStorageChannel,

		// 基础设施层 — 数据库连接
		persistence.NewDatabase,
		// 基础设施层 — Redis 连接
		persistence.NewRedisClient,

		// === 事件总线（Watermill Go Channel 实现） ===
		// 开发/测试环境使用 Go Channel，生产环境切换为 messaging.RedisStreamMQSet
		messaging.MQSet,

		// === 仓储基类 ===
		repo.NewRepositoryBase,

		// === KV 配置模块 ===
		repo.NewKvConfigRepository,
		repo.NewKvConfigReadRepository,
		wire.Bind(new(domainRepo.KvConfigRepository), new(*repo.KvConfigRepository)),
		wire.Bind(new(appRepo.KvConfigReadRepository), new(*repo.KvConfigReadRepository)),
		commandHandler.NewKvConfigCommandHandler,
		queryService.NewKvConfigQueryService,

		// === 字典模块 ===
		repo.NewDictTypeRepository,
		repo.NewDictTypeReadRepository,
		repo.NewDictCacheRepository,
		wire.Bind(new(domainRepo.DictTypeRepository), new(*repo.DictTypeRepository)),
		wire.Bind(new(appRepo.DictTypeReadRepository), new(*repo.DictTypeReadRepository)),
		wire.Bind(new(appRepo.DictCacheRepository), new(*repo.DictCacheRepository)),
		commandHandler.NewDictCommandHandler,
		queryService.NewDictQueryService,

		// === 认证模块 ===
		provideJWTConfig,
		repo.NewUserRepository,
		repo.NewUserReadRepository,
		repo.NewRoleReadRepository,
		repo.NewRoleRepository,
		repo.NewPermissionRepository,
		repo.NewPermissionReadRepository,
		wire.Bind(new(domainRepo.UserRepository), new(*repo.UserRepository)),
		wire.Bind(new(appRepo.UserReadRepository), new(*repo.UserReadRepository)),
		wire.Bind(new(appRepo.RoleReadRepository), new(*repo.RoleReadRepository)),
		wire.Bind(new(domainRepo.RoleRepository), new(*repo.RoleRepository)),
		wire.Bind(new(domainRepo.PermissionRepository), new(*repo.PermissionRepository)),
		wire.Bind(new(appRepo.PermissionReadRepository), new(*repo.PermissionReadRepository)),
		commandHandler.NewAuthCommandHandler,
		commandHandler.NewRbacCommandHandler,
		queryService.NewRoleQueryService,
		queryService.NewPermissionQueryService,
		queryService.NewUserQueryService,
		controller.NewAuthController,
		controller.NewRbacController,

		// === 审计日志模块 ===
		repo.NewAuditLogRepository,
		wire.Bind(new(appRepo.AuditLogReadRepository), new(*repo.AuditLogRepository)),
		queryService.NewAuditLogQueryService,
		controller.NewAuditController,
		controller.NewHealthController,
		controller.NewKvConfigController,
		controller.NewDictController,

		// === 文件存储模块 ===
		// 存储适配器（通过工厂创建）
		storageAdapter.NewFileStorage,
		// 文件记录仓储
		repo.NewFileRecordRepository,
		repo.NewFileRecordReadRepository,
		wire.Bind(new(domainRepo.FileRecordRepository), new(*repo.FileRecordRepository)),
		wire.Bind(new(appRepo.FileRecordReadRepository), new(*repo.FileRecordReadRepository)),
		commandHandler.NewFileCommandHandler,
		queryService.NewFileRecordQueryService,
		controller.NewFileController,

		// 应用组装
		NewApp,
	)
	return nil, nil, nil
}

// provideDatabaseConfig 从 Config 中提取 DatabaseConfig
func provideDatabaseConfig(cfg *config.Config) *config.DatabaseConfig {
	return &cfg.Database
}

// provideRedisConfig 从 Config 中提取 RedisConfig
func provideRedisConfig(cfg *config.Config) *config.RedisConfig {
	return &cfg.Redis
}

// provideJWTConfig 从 Config 中提取 JWTConfig
func provideJWTConfig(cfg *config.Config) *config.JWTConfig {
	return &cfg.JWT
}

// provideStorageConfig 从 Config 中提取 StorageConfig
func provideStorageConfig(cfg *config.Config) *config.StorageConfig {
	return &cfg.Storage
}

// provideStorageChannel 从 StorageConfig 中提取存储渠道标识
func provideStorageChannel(cfg *config.StorageConfig) string {
	return cfg.Driver
}