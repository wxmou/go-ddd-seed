package role

import "time"

// RolePermission 角色-权限关联实体
type RolePermission struct {
	ID           string
	RoleID       string
	PermissionID string
	CreatedAt    time.Time
}