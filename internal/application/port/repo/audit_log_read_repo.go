package repo

import (
	"context"
	"time"
)

// AuditLogReadDTO 审计日志读 DTO（CQRS 查询专用）
// 不包含请求/响应体，仅用于查询展示
type AuditLogReadDTO struct {
	ID           string    `json:"id"`
	OperatorID   string    `json:"operator_id"`
	OperatorName string    `json:"operator_name"`
	Action       string    `json:"action"`
	TargetType   string    `json:"target_type"`
	TargetID     string    `json:"target_id"`
	ClientIP     string    `json:"client_ip"`
	UserAgent    string    `json:"user_agent"`
	TraceID      string    `json:"trace_id"`
	CreatedAt    time.Time `json:"created_at"`
}

// AuditLogQuery 审计日志查询参数
type AuditLogQuery struct {
	OperatorID string    `json:"operator_id,omitempty"`
	Action     string    `json:"action,omitempty"`
	TargetType string    `json:"target_type,omitempty"`
	TargetID   string    `json:"target_id,omitempty"`
	StartTime  time.Time `json:"start_time,omitempty"`
	EndTime    time.Time `json:"end_time,omitempty"`
	Page       int       `json:"page"`
	PageSize   int       `json:"page_size"`
}

// AuditLogReadRepository 审计日志读仓储接口
type AuditLogReadRepository interface {
	// List 分页查询审计日志
	List(ctx context.Context, query AuditLogQuery) ([]*AuditLogReadDTO, int64, error)
}