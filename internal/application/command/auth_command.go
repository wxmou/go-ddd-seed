package command

// RegisterCommand 注册命令
type RegisterCommand struct {
	Username string `json:"username"`
	Password string `json:"password"`
	RealName string `json:"real_name"`
}

// LoginCommand 登录命令
type LoginCommand struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RefreshTokenCommand 刷新令牌命令
type RefreshTokenCommand struct {
	RefreshToken string `json:"refresh_token"`
}

// LogoutCommand 登出命令
type LogoutCommand struct {
	AccessToken string // 从请求头提取，不来自 JSON body
	UserID      string // 从当前登录用户获取
}

// ChangePasswordCommand 修改密码命令
type ChangePasswordCommand struct {
	UserID      string
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}