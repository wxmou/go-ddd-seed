package thirdPartyApi

import (
	"context"
	"io"
	"time"
)

// FileStorage 文件存储防腐层端口
// 定义在应用层，基础设施层实现适配器
// 文件存储属于纯技术支撑/外围服务，不计入领域规则，因此端口定义在应用层
type FileStorage interface {
	// Upload 上传文件，返回存储路径
	Upload(ctx context.Context, file io.Reader, path string) error
	// Download 下载文件
	Download(ctx context.Context, path string) (io.ReadCloser, error)
	// Delete 删除文件
	Delete(ctx context.Context, path string) error
	// GetURL 获取文件访问 URL（公开/临时签名）
	GetURL(ctx context.Context, path string, expiry time.Duration) (string, error)
	// Exists 文件是否存在
	Exists(ctx context.Context, path string) (bool, error)
}