package resp

// LoginResp 登录响应
type LoginResp struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	ExpiresIn    int64    `json:"expires_in"`
	User         UserInfo `json:"user"`
}

// UserInfo 用户信息响应
type UserInfo struct {
	ID          string   `json:"id"`
	Username    string   `json:"username"`
	RealName    string   `json:"real_name"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
}

// TokenResp 令牌响应（刷新时返回）
type TokenResp struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}