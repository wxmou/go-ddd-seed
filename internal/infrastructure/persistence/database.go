package persistence

import (
	"fmt"
	"github.com/go-ddd-seed/go-ddd-seed/internal/infrastructure/adapter/repo"
	"github.com/go-ddd-seed/go-ddd-seed/pkg/config"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlog "gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

// NewDatabase 创建数据库连接
func NewDatabase(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	var dial gorm.Dialector

	switch cfg.Driver {
	case "mysql":
		dial = mysql.Open(cfg.MySQLDSN())
	default:
		// 默认使用 PostgreSQL
		dial = postgres.Open(cfg.PostgresDSN())
	}

	db, err := gorm.Open(dial, &gorm.Config{
		Logger: gormlog.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			gormlog.Config{
				SlowThreshold:             time.Second,
				LogLevel:                  gormlog.Info,
				IgnoreRecordNotFoundError: true,
				Colorful:                  true,
			},
		),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// 连接池配置
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}

// AutoMigrate 自动迁移数据库表
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&repo.KvConfigGorm{},
		&repo.DictTypeGorm{},
		&repo.DictEntryGorm{},
		&repo.UserGorm{},
		&repo.RoleGorm{},
		&repo.PermissionGorm{},
		&repo.UserRoleGorm{},
		&repo.AuditLogGorm{},
		&repo.FileRecordGorm{},
	)
}
