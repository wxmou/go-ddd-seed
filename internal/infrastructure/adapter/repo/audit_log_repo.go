package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	appRepo "github.com/go-ddd-seed/go-ddd-seed/internal/application/port/repo"
	thirdPartyApi "github.com/go-ddd-seed/go-ddd-seed/internal/application/port/thirdPartyApi"
	"gorm.io/gorm"
)

// AuditLogReadGorm 审计日志读 GORM 模型（同写模型，用于分页查询）
type AuditLogReadGorm struct {
	ID           string    `gorm:"column:id"`
	OperatorID   string    `gorm:"column:operator_id"`
	OperatorName string    `gorm:"column:operator_name"`
	Action       string    `gorm:"column:action"`
	TargetType   string    `gorm:"column:target_type"`
	TargetID     string    `gorm:"column:target_id"`
	ClientIP     string    `gorm:"column:client_ip"`
	UserAgent    string    `gorm:"column:user_agent"`
	TraceID      string    `gorm:"column:trace_id"`
	CreatedAt    time.Time `gorm:"column:created_at"`
}

func (AuditLogReadGorm) TableName() string {
	return "audit_logs"
}

// 编译期接口检查
var _ thirdPartyApi.AuditLogRepository = (*AuditLogRepository)(nil)
var _ appRepo.AuditLogReadRepository = (*AuditLogRepository)(nil)

// AuditLogRepository 审计日志仓储（实现写 + 读）
type AuditLogRepository struct {
	DB *gorm.DB
}

// NewAuditLogRepository 创建审计日志仓储
func NewAuditLogRepository(db *gorm.DB) *AuditLogRepository {
	return &AuditLogRepository{DB: db}
}

// Save 写入审计日志
func (r *AuditLogRepository) Save(ctx context.Context, log *thirdPartyApi.AuditLogDTO) error {
	gormLog := &AuditLogGorm{
		ID:           uuid.New().String(),
		OperatorID:   log.OperatorID,
		OperatorName: log.OperatorName,
		Action:       log.Action,
		TargetType:   log.TargetType,
		TargetID:     log.TargetID,
		RequestBody:  log.RequestBody,
		ResponseBody: log.ResponseBody,
		ClientIP:     log.ClientIP,
		UserAgent:    log.UserAgent,
		TraceID:      log.TraceID,
		CreatedAt:    time.Now(),
	}
	return r.DB.WithContext(ctx).Create(gormLog).Error
}

// DeleteOlderThan 删除早于指定时间的记录
func (r *AuditLogRepository) DeleteOlderThan(ctx context.Context, before time.Time) (int64, error) {
	result := r.DB.WithContext(ctx).
		Where("created_at < ?", before).
		Delete(&AuditLogGorm{})
	return result.RowsAffected, result.Error
}

// List 分页查询审计日志
func (r *AuditLogRepository) List(ctx context.Context, query appRepo.AuditLogQuery) ([]*appRepo.AuditLogReadDTO, int64, error) {
	db := r.DB.WithContext(ctx).Model(&AuditLogReadGorm{})

	if query.OperatorID != "" {
		db = db.Where("operator_id = ?", query.OperatorID)
	}
	if query.Action != "" {
		db = db.Where("action = ?", query.Action)
	}
	if query.TargetType != "" {
		db = db.Where("target_type = ?", query.TargetType)
	}
	if query.TargetID != "" {
		db = db.Where("target_id = ?", query.TargetID)
	}
	if !query.StartTime.IsZero() {
		db = db.Where("created_at >= ?", query.StartTime)
	}
	if !query.EndTime.IsZero() {
		db = db.Where("created_at <= ?", query.EndTime)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 || query.PageSize > 100 {
		query.PageSize = 20
	}
	offset := (query.Page - 1) * query.PageSize

	var rows []AuditLogReadGorm
	if err := db.Order("created_at DESC").Offset(offset).Limit(query.PageSize).Find(&rows).Error; err != nil {
		return nil, 0, err
	}

	dtos := make([]*appRepo.AuditLogReadDTO, 0, len(rows))
	for _, r := range rows {
		dtos = append(dtos, &appRepo.AuditLogReadDTO{
			ID:           r.ID,
			OperatorID:   r.OperatorID,
			OperatorName: r.OperatorName,
			Action:       r.Action,
			TargetType:   r.TargetType,
			TargetID:     r.TargetID,
			ClientIP:     r.ClientIP,
			UserAgent:    r.UserAgent,
			TraceID:      r.TraceID,
			CreatedAt:    r.CreatedAt,
		})
	}
	return dtos, total, nil
}