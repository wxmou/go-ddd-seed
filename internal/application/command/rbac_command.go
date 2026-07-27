package command

// CreateRoleCommand 创建角色命令
type CreateRoleCommand struct {
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description"`
}

// UpdateRoleCommand 更新角色命令
type UpdateRoleCommand struct {
	ID          string // from URL param
	Name        string `json:"name"`
	Description string `json:"description"`
}

// DeleteRoleCommand 删除角色命令
type DeleteRoleCommand struct {
	ID string // from URL param
}

// EnableRoleCommand 启用角色命令
type EnableRoleCommand struct {
	ID string // from URL param
}

// DisableRoleCommand 禁用角色命令
type DisableRoleCommand struct {
	ID string // from URL param
}

// CreatePermissionCommand 创建权限命令
type CreatePermissionCommand struct {
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description"`
}

// UpdatePermissionCommand 更新权限命令
type UpdatePermissionCommand struct {
	ID          string // from URL param
	Name        string `json:"name"`
	Description string `json:"description"`
}

// DeletePermissionCommand 删除权限命令
type DeletePermissionCommand struct {
	ID string // from URL param
}

// AssignRolePermissionCommand 角色分配权限命令
type AssignRolePermissionCommand struct {
	RoleID       string   `json:"role_id"`
	PermissionID string   `json:"permission_id,omitempty"`
	// 批量分配时使用
	PermissionIDs []string `json:"permission_ids,omitempty"`
}

// RemoveRolePermissionCommand 角色移除权限命令
type RemoveRolePermissionCommand struct {
	RoleID       string // from path
	PermissionID string // from path
}

// AssignUserRoleCommand 用户分配角色命令
type AssignUserRoleCommand struct {
	UserID string `json:"user_id"`
	RoleID string `json:"role_id"`
}

// RemoveUserRoleCommand 用户移除角色命令
type RemoveUserRoleCommand struct {
	UserID string // from path
	RoleID string // from path
}
