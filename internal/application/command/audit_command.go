package command

// DeleteOldAuditLogsCommand 清理过期审计日志命令
type DeleteOldAuditLogsCommand struct {
	RetentionDays int // 保留天数
}