package main

import (
	"fmt"
	"github.com/go-ddd-seed/go-ddd-seed/pkg/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"net"
	"time"

	"context"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		fmt.Printf("❌ 加载配置失败: %v\n", err)
		return
	}
	fmt.Println("=== 数据库连接测试 ===")

	// 1. PostgreSQL
	testPostgres(cfg.Database)

	fmt.Println()

	// 2. Redis
	testRedis(cfg.Redis)

	fmt.Println()

	// 3. S3 兼容存储（MinIO 等，只检测端口连通性）
	testMinIO(cfg.Storage.S3)
}

func testPostgres(cfg config.DatabaseConfig) {
	dsn := cfg.PostgresDSN()
	fmt.Printf("[PostgreSQL] 尝试连接: %s:%d/%s ...\n", cfg.Host, cfg.Port, cfg.DBName)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Printf("❌ PostgreSQL 连接失败: %v\n", err)

		// 尝试 TCP 连通性检测
		addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
		conn, err := net.DialTimeout("tcp", addr, 3*time.Second)
		if err != nil {
			fmt.Printf("  └─ TCP 端口 %s 不可达: %v\n", addr, err)
		} else {
			conn.Close()
			fmt.Printf("  └─ TCP 端口可达，但数据库服务未正确响应\n")
		}
		return
	}

	sqlDB, err := db.DB()
	if err != nil {
		fmt.Printf("⚠️  获取 sql.DB 失败: %v\n", err)
		return
	}
	defer sqlDB.Close()

	if err := sqlDB.Ping(); err != nil {
		fmt.Printf("❌ PostgreSQL Ping 失败: %v\n", err)
		return
	}

	// 获取版本信息
	var version string
	db.Raw("SELECT version()").Scan(&version)
	fmt.Printf("✅ PostgreSQL 连接成功!\n")
	fmt.Printf("  └─ 版本: %s\n", version)
}

func testRedis(cfg config.RedisConfig) {
	fmt.Printf("[Redis] 尝试连接: %s:%d ...\n", cfg.Host, cfg.Port)

	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pong, err := rdb.Ping(ctx).Result()
	if err != nil {
		fmt.Printf("❌ Redis 连接失败: %v\n", err)

		addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
		conn, err := net.DialTimeout("tcp", addr, 3*time.Second)
		if err != nil {
			fmt.Printf("  └─ TCP 端口 %s 不可达: %v\n", addr, err)
		} else {
			conn.Close()
			fmt.Printf("  └─ TCP 端口可达，但 Redis 服务未正确响应\n")
		}
		return
	}

	fmt.Printf("✅ Redis 连接成功! Ping: %s\n", pong)
}

func testMinIO(cfg config.S3StorageConfig) {
	fmt.Printf("[MinIO] 尝试连接: %s ...\n", cfg.Endpoint)

	conn, err := net.DialTimeout("tcp", cfg.Endpoint, 3*time.Second)
	if err != nil {
		fmt.Printf("❌ MinIO TCP 连接失败: %v\n", err)
		return
	}
	conn.Close()
	fmt.Printf("✅ MinIO TCP 端口可达!\n")
	fmt.Printf("  └─ Endpoint: %s\n", cfg.Endpoint)
}
