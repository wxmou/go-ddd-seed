package role

import "github.com/go-ddd-seed/go-ddd-seed/internal/domain"

// Role 聚合错误码范围: 3100-3199
var (
	ErrRoleNameEmpty      = domain.NewDomainError(3100, "角色名称不能为空")
	ErrRoleCodeEmpty      = domain.NewDomainError(3101, "角色编码不能为空")
	ErrRoleCodeTooLong    = domain.NewDomainError(3102, "角色编码长度不能超过100")
	ErrRoleAlreadyEnabled = domain.NewDomainError(3103, "角色已启用")
	ErrRoleAlreadyDisabled = domain.NewDomainError(3104, "角色已禁用")
)