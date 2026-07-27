package repo

import (
	"context"
	"time"
)

// FileRecordReadDTO 文件记录读 DTO（CQRS 读模型）
type FileRecordReadDTO struct {
	ID             string    `json:"id"`
	FileName       string    `json:"file_name"`
	StoragePath    string    `json:"storage_path"`
	Size           int64     `json:"size"`
	MIMEType       string    `json:"mime_type"`
	StorageChannel string    `json:"storage_channel"`
	MD5Hash        string    `json:"md5_hash"`
	AttachType     string    `json:"attach_type,omitempty"`
	AttachID       string    `json:"attach_id,omitempty"`
	UploaderID     string    `json:"uploader_id"`
	ThumbnailPath  string    `json:"thumbnail_path,omitempty"`
	ThumbnailURL   string    `json:"thumbnail_url,omitempty"`
	AccessURL      string    `json:"access_url,omitempty"`
	Status         int       `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// FileRecordQuery 文件记录查询条件
type FileRecordQuery struct {
	AttachType string `json:"attach_type"` // 关联业务类型
	AttachID   string `json:"attach_id"`   // 关联业务 ID
	UploaderID string `json:"uploader_id"` // 上传者 ID
	Status     *int   `json:"status"`      // 状态筛选
	Page       int    `json:"page"`        // 页码
	PageSize   int    `json:"page_size"`   // 每页数量
}

// FileRecordReadRepository 文件记录读仓储接口
// 定义在应用层端口，基础设施层实现
type FileRecordReadRepository interface {
	// FindByID 按 ID 查询文件记录
	FindByID(ctx context.Context, id string) (*FileRecordReadDTO, error)
	// FindByAttach 按业务对象查询关联文件
	FindByAttach(ctx context.Context, attachType, attachID string) ([]*FileRecordReadDTO, error)
	// List 分页列表查询
	List(ctx context.Context, query FileRecordQuery) ([]*FileRecordReadDTO, int64, error)
}