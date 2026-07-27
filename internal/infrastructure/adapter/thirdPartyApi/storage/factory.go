package storage

import (
	"fmt"

	appApi "github.com/go-ddd-seed/go-ddd-seed/internal/application/port/thirdPartyApi"
	"github.com/go-ddd-seed/go-ddd-seed/pkg/config"
)

// NewFileStorage 根据配置创建对应的存储适配器
func NewFileStorage(cfg *config.StorageConfig) (appApi.FileStorage, error) {
	switch cfg.Driver {
	case "local":
		return NewLocalStorage(cfg.Local.Path, cfg.Local.BaseURL), nil
	case "s3":
		return NewMinIOStorage(cfg.S3.Endpoint, cfg.S3.AccessKey, cfg.S3.SecretKey, cfg.S3.Bucket, cfg.S3.UseSSL)
	default:
		return nil, fmt.Errorf("不支持的存储渠道: %s", cfg.Driver)
	}
}