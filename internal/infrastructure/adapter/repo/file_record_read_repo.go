package repo

import (
	"context"
	"time"

	"gorm.io/gorm"

	appRepo "github.com/go-ddd-seed/go-ddd-seed/internal/application/port/repo"
)

// Ensure interface compliance
var _ appRepo.FileRecordReadRepository = (*FileRecordReadRepository)(nil)

// FileRecordReadGorm 文件记录读 GORM 模型
type FileRecordReadGorm struct {
	ID             string    `gorm:"column:id;primaryKey;type:varchar(36)"`
	FileName       string    `gorm:"column:file_name;type:varchar(255);not null"`
	StoragePath    string    `gorm:"column:storage_path;type:varchar(1024);not null"`
	Size           int64     `gorm:"column:size;not null"`
	MIMEType       string    `gorm:"column:mime_type;type:varchar(127)"`
	StorageChannel string    `gorm:"column:storage_channel;type:varchar(32)"`
	MD5Hash        string    `gorm:"column:md5_hash;type:varchar(64)"`
	AttachType     string    `gorm:"column:attach_type;type:varchar(64)"`
	AttachID       string    `gorm:"column:attach_id;type:varchar(36)"`
	UploaderID     string    `gorm:"column:uploader_id;type:varchar(36)"`
	ThumbnailPath  string    `gorm:"column:thumbnail_path;type:varchar(1024)"`
	Status         int       `gorm:"column:status;default:1"`
	CreatedAt      time.Time `gorm:"column:created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at"`
}

// TableName 表名
func (FileRecordReadGorm) TableName() string {
	return "file_records"
}

// FileRecordReadRepository 文件记录读仓储实现
type FileRecordReadRepository struct {
	DB *gorm.DB
}

// NewFileRecordReadRepository 创建文件记录读仓储
func NewFileRecordReadRepository(db *gorm.DB) *FileRecordReadRepository {
	return &FileRecordReadRepository{DB: db}
}

// toDTO 从 GORM 模型转换到 DTO
func (r *FileRecordReadRepository) toDTO(gorm *FileRecordReadGorm) *appRepo.FileRecordReadDTO {
	return &appRepo.FileRecordReadDTO{
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

// FindByID 按 ID 查询文件记录
func (r *FileRecordReadRepository) FindByID(ctx context.Context, id string) (*appRepo.FileRecordReadDTO, error) {
	var gormModel FileRecordReadGorm
	err := r.DB.WithContext(ctx).Where("id = ?", id).First(&gormModel).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return r.toDTO(&gormModel), nil
}

// FindByAttach 按业务对象查询关联文件
func (r *FileRecordReadRepository) FindByAttach(ctx context.Context, attachType, attachID string) ([]*appRepo.FileRecordReadDTO, error) {
	var gormModels []FileRecordReadGorm
	err := r.DB.WithContext(ctx).
		Where("attach_type = ? AND attach_id = ?", attachType, attachID).
		Order("created_at ASC").
		Find(&gormModels).Error
	if err != nil {
		return nil, err
	}

	result := make([]*appRepo.FileRecordReadDTO, 0, len(gormModels))
	for i := range gormModels {
		result = append(result, r.toDTO(&gormModels[i]))
	}
	return result, nil
}

// List 分页列表查询
func (r *FileRecordReadRepository) List(ctx context.Context, query appRepo.FileRecordQuery) ([]*appRepo.FileRecordReadDTO, int64, error) {
	db := r.DB.WithContext(ctx).Model(&FileRecordReadGorm{})

	// 条件筛选
	if query.AttachType != "" {
		db = db.Where("attach_type = ?", query.AttachType)
	}
	if query.AttachID != "" {
		db = db.Where("attach_id = ?", query.AttachID)
	}
	if query.UploaderID != "" {
		db = db.Where("uploader_id = ?", query.UploaderID)
	}
	if query.Status != nil {
		db = db.Where("status = ?", *query.Status)
	}

	// 总数
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}
	offset := (query.Page - 1) * query.PageSize

	var gormModels []FileRecordReadGorm
	err := db.Order("created_at DESC").
		Offset(offset).
		Limit(query.PageSize).
		Find(&gormModels).Error
	if err != nil {
		return nil, 0, err
	}

	result := make([]*appRepo.FileRecordReadDTO, 0, len(gormModels))
	for i := range gormModels {
		result = append(result, r.toDTO(&gormModels[i]))
	}
	return result, total, nil
}