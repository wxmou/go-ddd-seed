package config

import (
	"github.com/spf13/viper"
	"strings"
)

// Config 应用配置
type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Redis     RedisConfig     `mapstructure:"redis"`
	Storage   StorageConfig   `mapstructure:"storage"`
	Scheduler SchedulerConfig `mapstructure:"scheduler"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	Log       LogConfig       `mapstructure:"log"`
}

// ServerConfig 服务配置
type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver   string `mapstructure:"driver"` // postgres / mysql
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

// DSN 生成数据库连接字符串
func (d DatabaseConfig) DSN() string {
	switch d.Driver {
	case "mysql":
		return d.MySQLDSN()
	default:
		return d.PostgresDSN()
	}
}

// PostgresDSN PostgreSQL 连接字符串
func (d DatabaseConfig) PostgresDSN() string {
	return "host=" + d.Host +
		" user=" + d.User +
		" password=" + d.Password +
		" dbname=" + d.DBName +
		" port=" + itoa(d.Port) +
		" sslmode=" + d.SSLMode +
		" TimeZone=Asia/Shanghai"
}

// MySQLDSN MySQL 连接字符串
func (d DatabaseConfig) MySQLDSN() string {
	return d.User + ":" + d.Password +
		"@tcp(" + d.Host + ":" + itoa(d.Port) + ")/" + d.DBName +
		"?charset=utf8mb4&parseTime=True&loc=Local"
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// Addr Redis 地址
func (r RedisConfig) Addr() string {
	return r.Host + ":" + itoa(r.Port)
}

// StorageConfig 存储配置
type StorageConfig struct {
	// Driver 存储渠道驱动：local / s3（MinIO 也使用 s3，兼容 S3 协议）
	Driver string            `mapstructure:"driver"`
	Local  LocalStorageConfig `mapstructure:"local"`
	S3     S3StorageConfig    `mapstructure:"s3"`
}

// LocalStorageConfig 本地存储配置
type LocalStorageConfig struct {
	Path    string `mapstructure:"path"`
	BaseURL string `mapstructure:"baseURL"`
}

// S3StorageConfig S3 兼容存储配置（MinIO 兼容 S3 协议，复用此配置）
type S3StorageConfig struct {
	Endpoint  string `mapstructure:"endpoint"`
	AccessKey string `mapstructure:"accessKey"`
	SecretKey string `mapstructure:"secretKey"`
	Bucket    string `mapstructure:"bucket"`
	UseSSL    bool   `mapstructure:"useSSL"`
}

// JWTConfig JWT 配置
type JWTConfig struct {
	Secret             string `mapstructure:"secret"`
	Expiration         int    `mapstructure:"expiration"`         // Access Token 过期时间（小时）
	RefreshExpiration  int    `mapstructure:"refreshExpiration"`  // Refresh Token 过期时间（小时）
	DefaultRole        string `mapstructure:"defaultRole"`        // 注册默认角色 code
}

// LogConfig 日志配置
type LogConfig struct {
	Level         string `mapstructure:"level"`         // debug, info, warn, error
	Format        string `mapstructure:"format"`        // json, text
	SlowThreshold int64  `mapstructure:"slowThreshold"` // 慢请求阈值（毫秒），默认 500
}

// SchedulerConfig 定时任务调度器配置
type SchedulerConfig struct {
	// Enabled 是否启用调度器
	Enabled bool `mapstructure:"enabled"`
	// AuditCleanupCron 审计日志清理 cron 表达式（秒级）
	AuditCleanupCron string `mapstructure:"audit_cleanup_cron"`
	// AuditRetentionDays 审计日志保留天数
	AuditRetentionDays int `mapstructure:"audit_retention_days"`
	// JobTimeout 任务超时时间（秒）
	JobTimeout int `mapstructure:"job_timeout"`
}

// Load 加载配置文件
func Load(configPath string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(configPath)

	// 环境变量覆盖
	v.SetEnvPrefix("APP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	s := ""
	for i > 0 {
		s = string(rune('0'+i%10)) + s
		i /= 10
	}
	return s
}