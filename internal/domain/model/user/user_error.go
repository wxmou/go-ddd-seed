package user

import "github.com/go-ddd-seed/go-ddd-seed/internal/domain"

// User 聚合错误码范围: 3000-3099
var (
	ErrUsernameEmpty       = domain.NewDomainError(3000, "用户名不能为空")
	ErrUsernameTooLong     = domain.NewDomainError(3001, "用户名长度不能超过100")
	ErrPasswordTooShort    = domain.NewDomainError(3002, "密码长度不能少于6位")
	ErrPasswordTooLong     = domain.NewDomainError(3003, "密码长度不能超过128")
	ErrRealNameTooLong     = domain.NewDomainError(3004, "真实姓名长度不能超过100")
	ErrEmailTooLong        = domain.NewDomainError(3005, "邮箱长度不能超过200")
	ErrPhoneTooLong        = domain.NewDomainError(3006, "手机号长度不能超过50")
	ErrUserAlreadyEnabled  = domain.NewDomainError(3007, "用户已启用")
	ErrUserAlreadyDisabled = domain.NewDomainError(3008, "用户已禁用")
)