package req

// AuditLogQueryReq 审计日志查询请求
type AuditLogQueryReq struct {
	OperatorID string `form:"operator_id"`
	Action     string `form:"action"`       // create / update / delete
	TargetType string `form:"target_type"`
	TargetID   string `form:"target_id"`
	StartTime  string `form:"start_time"`   // RFC3339
	EndTime    string `form:"end_time"`     // RFC3339
	Page       int    `form:"page"`         // 默认 1
	PageSize   int    `form:"page_size"`    // 默认 20，最大 100
}
