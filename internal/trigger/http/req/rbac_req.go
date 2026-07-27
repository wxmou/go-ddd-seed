package req

// CreateRoleReq 创建角色请求
type CreateRoleReq struct {
	Name        string `json:"name" binding:"required,min=1,max=100"`
	Code        string `json:"code" binding:"required,min=1,max=100,alphanum"`
	Description string `json:"description" binding:"max=500"`
}

// UpdateRoleReq 更新角色请求
type UpdateRoleReq struct {
	Name        string `json:"name" binding:"required,min=1,max=100"`
	Description string `json:"description" binding:"max=500"`
}

// CreatePermissionReq 创建权限请求
type CreatePermissionReq struct {
	Name        string `json:"name" binding:"required,min=1,max=200"`
	Code        string `json:"code" binding:"required,min=1,max=200,alphanum"`
	Description string `json:"description" binding:"max=500"`
}

// UpdatePermissionReq 更新权限请求
type UpdatePermissionReq struct {
	Name        string `json:"name" binding:"required,min=1,max=200"`
	Description string `json:"description" binding:"max=500"`
}

// AssignPermissionReq 分配权限请求
type AssignPermissionReq struct {
	PermissionIDs []string `json:"permission_ids" binding:"required,min=1"`
}

// AssignRoleReq 分配角色请求
type AssignRoleReq struct {
	RoleID string `json:"role_id" binding:"required"`
}