package thirdPartyApi

import (
	"context"
	"time"
)

// AuditLogDTO 审计日志写数据（写仓储入参）
// 包含请求/响应体，仅用于写入场景
type AuditLogDTO struct {
	ID           string    `json:"id"`
	OperatorID   string    `json:"operator_id"`
	OperatorName string    `json:"operator_name"`
	Action       string    `json:"action"`      // create / update / delete
	TargetType   string    `json:"target_type"` // role / permission / user / dict_type / kv_config
	TargetID     string    `json:"target_id"`
	RequestBody  string    `json:"request_body,omitempty"`
	ResponseBody string    `json:"response_body,omitempty"`
	ClientIP     string    `json:"client_ip"`
	UserAgent    string    `json:"user_agent"`
	TraceID      string    `json:"trace_id"`
	CreatedAt    time.Time `json:"created_at"`
}

// AuditLogRepository 审计日志写仓储
// append-only：只追加，不修改不删除
type AuditLogRepository interface {
	Save(ctx context.Context, log *AuditLogDTO) error
	// DeleteOlderThan 删除早于指定时间的记录（用于定时清理）
	DeleteOlderThan(ctx context.Context, before time.Time) (int64, error)
}