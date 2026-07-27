package repo

import (
	"context"
	"time"

	"github.com/go-ddd-seed/go-ddd-seed/internal/domain"
	"gorm.io/gorm"

	domainRepo "github.com/go-ddd-seed/go-ddd-seed/internal/domain/port/repo"
	"github.com/go-ddd-seed/go-ddd-seed/internal/domain/model/file_record"
)

// Ensure interface compliance
var _ domainRepo.FileRecordRepository = (*FileRecordRepository)(nil)

// FileRecordGorm 文件记录 GORM 模型
type FileRecordGorm struct {
	ID             string    `gorm:"column:id;primaryKey;type:varchar(36)"`
	FileName       string    `gorm:"column:file_name;type:varchar(255);not null"`
	StoragePath    string    `gorm:"column:storage_path;type:varchar(1024);not null"`
	Size           int64     `gorm:"column:size;not null"`
	MIMEType       string    `gorm:"column:mime_type;type:varchar(127)"`
	StorageChannel string    `gorm:"column:storage_channel;type:varchar(32)"`
	MD5Hash        string    `gorm:"column:md5_hash;type:varchar(64)"`
	AttachType     string    `gorm:"column:attach_type;type:varchar(64);index:idx_file_attach"`
	AttachID       string    `gorm:"column:attach_id;type:varchar(36);index:idx_file_attach"`
	UploaderID     string    `gorm:"column:uploader_id;type:varchar(36);index:idx_file_uploader"`
	ThumbnailPath  string    `gorm:"column:thumbnail_path;type:varchar(1024)"`
	Status         int       `gorm:"column:status;default:1"`
	CreatedAt      time.Time `gorm:"column:created_at;index:idx_file_created"`
	UpdatedAt      time.Time `gorm:"column:updated_at"`
}

// TableName 表名
func (FileRecordGorm) TableName() string {
	return "file_records"
}

// fromFileRecordDomain 从领域聚合转换
func fromFileRecordDomain(fr *file_record.FileRecord) *FileRecordGorm {
	return &FileRecordGorm{
		ID:             fr.ID,
		FileName:       fr.FileName,
		StoragePath:    fr.StoragePath,
		Size:           fr.Size,
		MIMEType:       fr.MIMEType,
		StorageChannel: fr.StorageChannel,
		MD5Hash:        fr.MD5Hash,
		AttachType:     fr.AttachType,
		AttachID:       fr.AttachID,
		UploaderID:     fr.UploaderID,
		ThumbnailPath:  fr.ThumbnailPath,
		Status:         fr.Status,
		CreatedAt:      fr.CreatedAt,
		UpdatedAt:      fr.UpdatedAt,
	}
}

// toFileRecordDomain 从 GORM 模型转换回领域聚合
func toFileRecordDomain(gorm *FileRecordGorm) *file_record.FileRecord {
	return &file_record.FileRecord{
		ID:             gorm.ID,
		FileName:       gorm.FileName,
		StoragePath:    gorm.StoragePath,
		Size:           gorm.Size,
		MIMEType:       gorm.MIMEType,
		StorageChannel: gorm.StorageChannel,
		MD5Hash:        gorm.MD5Hash,
		AttachType:     gorm.AttachType,
		AttachID:       gorm.AttachID,
		UploaderID:     gorm.UploaderID,
		ThumbnailPath:  gorm.ThumbnailPath,
		Status:         gorm.Status,
		CreatedAt:      gorm.CreatedAt,
		UpdatedAt:      gorm.UpdatedAt,
	}
}

// FileRecordRepository GORM 仓储实现（命令侧）
type FileRecordRepository struct {
	RepositoryBase
}

// NewFileRecordRepository 创建文件记录仓储
func NewFileRecordRepository(base RepositoryBase) *FileRecordRepository {
	return &FileRecordRepository{RepositoryBase: base}
}

// Save 保存文件记录，自动发布领域事件
func (r *FileRecordRepository) Save(ctx context.Context, record *file_record.FileRecord) error {
	return r.SaveWithEvents(ctx, record, func(tx *gorm.DB) error {
		gormModel := fromFileRecordDomain(record)
		return tx.Save(gormModel).Error
	})
}

// FindByID 按 ID 加载文件记录
func (r *FileRecordRepository) FindByID(ctx context.Context, id string) (*file_record.FileRecord, error) {
	var gormModel FileRecordGorm
	err := r.DB.WithContext(ctx).Where("id = ?", id).First(&gormModel).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrRecordNotFound
		}
		return nil, err
	}

	return toFileRecordDomain(&gormModel), nil
}

// FindByAttach 按业务对象查询关联文件
func (r *FileRecordRepository) FindByAttach(ctx context.Context, attachType, attachID string) ([]*file_record.FileRecord, error) {
	var gormModels []FileRecordGorm
	err := r.DB.WithContext(ctx).
		Where("attach_type = ? AND attach_id = ?", attachType, attachID).
		Order("created_at ASC").
		Find(&gormModels).Error
	if err != nil {
		return nil, err
	}

	result := make([]*file_record.FileRecord, 0, len(gormModels))
	for i := range gormModels {
		result = append(result, toFileRecordDomain(&gormModels[i]))
	}
	return result, nil
}

// Delete 删除文件记录
func (r *FileRecordRepository) Delete(ctx context.Context, id string) error {
	result := r.DB.WithContext(ctx).Where("id = ?", id).Delete(&FileRecordGorm{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrRecordNotFound
	}
	return nil
}