package resp

// AuditLogItem 审计日志条目
type AuditLogItem struct {
	ID           string `json:"id"`
	OperatorID   string `json:"operator_id"`
	OperatorName string `json:"operator_name"`
	Action       string `json:"action"`
	TargetType   string `json:"target_type"`
	TargetID     string `json:"target_id"`
	ClientIP     string `json:"client_ip"`
	UserAgent    string `json:"user_agent"`
	TraceID      string `json:"trace_id"`
	CreatedAt    string `json:"created_at"`
}

// AuditLogListResp 审计日志列表响应
type AuditLogListResp struct {
	Items []AuditLogItem `json:"items"`
	Total int64          `json:"total"`
	Page  int            `json:"page"`
}