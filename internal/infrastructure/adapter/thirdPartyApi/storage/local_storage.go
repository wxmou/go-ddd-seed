package storage

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"time"

	appApi "github.com/go-ddd-seed/go-ddd-seed/internal/application/port/thirdPartyApi"
)

// LocalStorage 本地文件存储适配器
type LocalStorage struct {
	basePath string // 文件存储根目录
	baseURL  string // 公开访问 URL 前缀
}

// 编译期检查接口实现
var _ appApi.FileStorage = (*LocalStorage)(nil)

// NewLocalStorage 创建本地存储适配器
func NewLocalStorage(basePath, baseURL string) *LocalStorage {
	// 确保目录存在
	os.MkdirAll(basePath, 0755)
	return &LocalStorage{basePath: basePath, baseURL: baseURL}
}

// Upload 上传文件到本地存储
func (s *LocalStorage) Upload(ctx context.Context, file io.Reader, path string) error {
	fullPath := filepath.Join(s.basePath, path)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return err
	}

	dst, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	return err
}

// Download 从本地存储下载文件
func (s *LocalStorage) Download(ctx context.Context, path string) (io.ReadCloser, error) {
	return os.Open(filepath.Join(s.basePath, path))
}

// Delete 从本地存储删除文件
func (s *LocalStorage) Delete(ctx context.Context, path string) error {
	return os.Remove(filepath.Join(s.basePath, path))
}

// GetURL 获取文件访问 URL
func (s *LocalStorage) GetURL(ctx context.Context, path string, expiry time.Duration) (string, error) {
	return s.baseURL + "/" + path, nil
}

// Exists 检查文件是否存在
func (s *LocalStorage) Exists(ctx context.Context, path string) (bool, error) {
	_, err := os.Stat(filepath.Join(s.basePath, path))
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, err
}