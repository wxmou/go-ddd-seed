package repo

import "time"

// AuditLogGorm 审计日志 GORM 模型
type AuditLogGorm struct {
	ID           string    `gorm:"column:id;primaryKey;type:varchar(36)"`
	OperatorID   string    `gorm:"column:operator_id;type:varchar(36);not null;index:idx_audit_operator"`
	OperatorName string    `gorm:"column:operator_name;type:varchar(100);not null;default:''"`
	Action       string    `gorm:"column:action;type:varchar(32);not null;index:idx_audit_action"`
	TargetType   string    `gorm:"column:target_type;type:varchar(64);not null;index:idx_audit_target"`
	TargetID     string    `gorm:"column:target_id;type:varchar(36);not null;default:''"`
	RequestBody  string    `gorm:"column:request_body;type:text;not null;default:''"`
	ResponseBody string    `gorm:"column:response_body;type:text;not null;default:''"`
	ClientIP     string    `gorm:"column:client_ip;type:varchar(45);not null;default:''"`
	UserAgent    string    `gorm:"column:user_agent;type:varchar(512);not null;default:''"`
	TraceID      string    `gorm:"column:trace_id;type:varchar(36);not null;default:'';index"`
	CreatedAt    time.Time `gorm:"column:created_at;index:idx_audit_created"`
}

func (AuditLogGorm) TableName() string {
	return "audit_logs"
}
