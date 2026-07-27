package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	appApi "github.com/go-ddd-seed/go-ddd-seed/internal/application/port/thirdPartyApi"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinIOStorage MinIO/S3 文件存储适配器
type MinIOStorage struct {
	client  *minio.Client
	bucket  string
	baseURL string // 公开访问 URL 前缀（可选）
}

// 编译期检查接口实现
var _ appApi.FileStorage = (*MinIOStorage)(nil)

// NewMinIOStorage 创建 MinIO 存储适配器
func NewMinIOStorage(endpoint, accessKey, secretKey, bucket string, useSSL bool) (*MinIOStorage, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("创建 MinIO 客户端失败: %w", err)
	}

	// 确保 bucket 存在
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("检查 MinIO bucket 失败: %w", err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("创建 MinIO bucket 失败: %w", err)
		}
	}

	return &MinIOStorage{client: client, bucket: bucket}, nil
}

// Upload 上传文件到 MinIO
func (s *MinIOStorage) Upload(ctx context.Context, file io.Reader, path string) error {
	_, err := s.client.PutObject(ctx, s.bucket, path, file, -1, minio.PutObjectOptions{})
	return err
}

// Download 从 MinIO 下载文件
func (s *MinIOStorage) Download(ctx context.Context, path string) (io.ReadCloser, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, path, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	return obj, nil
}

// Delete 从 MinIO 删除文件
func (s *MinIOStorage) Delete(ctx context.Context, path string) error {
	return s.client.RemoveObject(ctx, s.bucket, path, minio.RemoveObjectOptions{})
}

// GetURL 获取文件访问 URL（临时签名 URL）
func (s *MinIOStorage) GetURL(ctx context.Context, path string, expiry time.Duration) (string, error) {
	reqParams := url.Values{}
	presignedURL, err := s.client.PresignedGetObject(ctx, s.bucket, path, expiry, reqParams)
	if err != nil {
		return "", err
	}
	return presignedURL.String(), nil
}

// Exists 检查文件是否存在
func (s *MinIOStorage) Exists(ctx context.Context, path string) (bool, error) {
	_, err := s.client.StatObject(ctx, s.bucket, path, minio.StatObjectOptions{})
	if err != nil {
		errResp := minio.ToErrorResponse(err)
		if errResp.Code == "NoSuchKey" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}