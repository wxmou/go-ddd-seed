package file_record

import (
	"time"

	"github.com/go-ddd-seed/go-ddd-seed/internal/domain"
	"github.com/go-ddd-seed/go-ddd-seed/internal/domain/event"
	"github.com/go-ddd-seed/go-ddd-seed/internal/domain/model"
)

// FileRecordStatus 文件记录状态
const (
	FileRecordStatusDeleted  = 0
	FileRecordStatusNormal   = 1
)

// 领域事件名称
const (
	EventFileUploaded      = "file.uploaded"
	EventFileDeleted       = "file.deleted"
	EventThumbnailGenerated = "file.thumbnail.generated"
)

// FileUploadedEvent 文件上传事件
type FileUploadedEvent struct {
	event.BaseEvent
	AggregateID    string `json:"aggregate_id"`
	FileName       string `json:"file_name"`
	StoragePath    string `json:"storage_path"`
	StorageChannel string `json:"storage_channel"`
	MIMEType       string `json:"mime_type"`
}

// NewFileUploadedEvent 创建文件上传事件
func NewFileUploadedEvent(aggregateID, fileName, storagePath, storageChannel, mimeType string) FileUploadedEvent {
	return FileUploadedEvent{
		BaseEvent: event.BaseEvent{
			At:   time.Now(),
			Name: EventFileUploaded,
		},
		AggregateID:    aggregateID,
		FileName:       fileName,
		StoragePath:    storagePath,
		StorageChannel: storageChannel,
		MIMEType:       mimeType,
	}
}

// FileDeletedEvent 文件删除事件
type FileDeletedEvent struct {
	event.BaseEvent
	AggregateID string `json:"aggregate_id"`
}

// NewFileDeletedEvent 创建文件删除事件
func NewFileDeletedEvent(aggregateID string) FileDeletedEvent {
	return FileDeletedEvent{
		BaseEvent: event.BaseEvent{
			At:   time.Now(),
			Name: EventFileDeleted,
		},
		AggregateID: aggregateID,
	}
}

// ThumbnailGeneratedEvent 缩略图生成事件
type ThumbnailGeneratedEvent struct {
	event.BaseEvent
	AggregateID   string `json:"aggregate_id"`
	ThumbnailPath string `json:"thumbnail_path"`
}

// NewThumbnailGeneratedEvent 创建缩略图生成事件
func NewThumbnailGeneratedEvent(aggregateID, thumbnailPath string) ThumbnailGeneratedEvent {
	return ThumbnailGeneratedEvent{
		BaseEvent: event.BaseEvent{
			At:   time.Now(),
			Name: EventThumbnailGenerated,
		},
		AggregateID:   aggregateID,
		ThumbnailPath: thumbnailPath,
	}
}

// FileRecord 文件记录聚合根
type FileRecord struct {
	model.AggregateRoot
	ID             string
	FileName       string    // 原始文件名
	StoragePath    string    // 存储后端路径
	Size           int64     // 文件大小（字节）
	MIMEType       string    // MIME 类型
	StorageChannel string    // 存储渠道标识
	MD5Hash        string    // 文件 MD5 校验
	AttachType     string    // 关联业务类型
	AttachID       string    // 关联业务 ID
	UploaderID     string    // 上传者 ID
	ThumbnailPath  string    // 缩略图存储路径
	Status         int       // 1=正常, 0=已删除
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// NewFileRecord 创建新的文件记录（构造器完成字段校验、填充默认值）
func NewFileRecord(id, fileName, storagePath, mimeType, storageChannel, md5Hash, uploaderID string, size int64) (*FileRecord, error) {
	if fileName == "" {
		return nil, ErrFileNameEmpty
	}
	if storagePath == "" {
		return nil, ErrStoragePathEmpty
	}
	if uploaderID == "" {
		return nil, ErrUploaderIDEmpty
	}

	now := time.Now()
	fr := &FileRecord{
		ID:             id,
		FileName:       fileName,
		StoragePath:    storagePath,
		Size:           size,
		MIMEType:       mimeType,
		StorageChannel: storageChannel,
		MD5Hash:        md5Hash,
		UploaderID:     uploaderID,
		Status:         FileRecordStatusNormal,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	fr.AddDomainEvent(NewFileUploadedEvent(id, fileName, storagePath, storageChannel, mimeType))
	return fr, nil
}

// Delete 软删除文件记录
func (fr *FileRecord) Delete() error {
	if fr.Status == FileRecordStatusDeleted {
		return ErrFileRecordAlreadyDeleted
	}
	fr.Status = FileRecordStatusDeleted
	fr.UpdatedAt = time.Now()
	fr.AddDomainEvent(NewFileDeletedEvent(fr.ID))
	return nil
}

// Associate 关联业务对象
func (fr *FileRecord) Associate(attachType, attachID string) error {
	if attachType == "" {
		return domain.NewDomainError(ErrCodeAttachTypeEmpty, "关联业务类型不能为空")
	}
	if attachID == "" {
		return domain.NewDomainError(ErrCodeAttachIDEmpty, "关联业务ID不能为空")
	}
	fr.AttachType = attachType
	fr.AttachID = attachID
	fr.UpdatedAt = time.Now()
	return nil
}

// SetThumbnail 设置缩略图路径
func (fr *FileRecord) SetThumbnail(path string) {
	fr.ThumbnailPath = path
	fr.UpdatedAt = time.Now()
	fr.AddDomainEvent(NewThumbnailGeneratedEvent(fr.ID, path))
}

// IsImage 是否为图片类型文件
func (fr *FileRecord) IsImage() bool {
	return len(fr.MIMEType) >= 6 && fr.MIMEType[:6] == "image/"
}

// IsDeleted 是否已删除
func (fr *FileRecord) IsDeleted() bool {
	return fr.Status == FileRecordStatusDeleted
}

// FileRecord 聚合错误码范围: 4000-4099
const (
	ErrCodeFileNameEmpty    = 4000
	ErrCodeStoragePathEmpty = 4001
	ErrCodeUploaderIDEmpty  = 4002
	ErrCodeAttachTypeEmpty  = 4003
	ErrCodeAttachIDEmpty    = 4004
	ErrCodeAlreadyDeleted   = 4005
)

// 领域错误
var (
	ErrFileNameEmpty           = domain.NewDomainError(ErrCodeFileNameEmpty, "文件名不能为空")
	ErrStoragePathEmpty        = domain.NewDomainError(ErrCodeStoragePathEmpty, "存储路径不能为空")
	ErrUploaderIDEmpty         = domain.NewDomainError(ErrCodeUploaderIDEmpty, "上传者ID不能为空")
	ErrFileRecordAlreadyDeleted = domain.NewDomainError(ErrCodeAlreadyDeleted, "文件记录已删除")
)