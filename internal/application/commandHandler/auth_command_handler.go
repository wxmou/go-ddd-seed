package commandHandler

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	appRepo "github.com/go-ddd-seed/go-ddd-seed/internal/application/port/repo"
	"github.com/go-ddd-seed/go-ddd-seed/internal/application/command"
	domainRepo "github.com/go-ddd-seed/go-ddd-seed/internal/domain/port/repo"
	"github.com/go-ddd-seed/go-ddd-seed/internal/domain/model/user"
	"github.com/go-ddd-seed/go-ddd-seed/pkg/auth"
	"github.com/go-ddd-seed/go-ddd-seed/pkg/config"
	"github.com/go-ddd-seed/go-ddd-seed/pkg/crypto"
	appErrors "github.com/go-ddd-seed/go-ddd-seed/pkg/errors"
)

// AuthCommandHandler 认证命令处理器
// 注意：认证处理器并非纯写操作，Login/RefreshToken 需要跨聚合查询角色权限，
// 因此 userRead/roleRead 作为特例保留，用于认证流程中的只读查询。
type AuthCommandHandler struct {
	userRepo domainRepo.UserRepository
	userRead appRepo.UserReadRepository
	roleRead appRepo.RoleReadRepository
	rdb      *redis.Client
	jwtCfg   *config.JWTConfig
}

// NewAuthCommandHandler 创建认证命令处理器
func NewAuthCommandHandler(
	userRepo domainRepo.UserRepository,
	userRead appRepo.UserReadRepository,
	roleRead appRepo.RoleReadRepository,
	rdb *redis.Client,
	jwtCfg *config.JWTConfig,
) *AuthCommandHandler {
	return &AuthCommandHandler{
		userRepo: userRepo,
		userRead: userRead,
		roleRead: roleRead,
		rdb:      rdb,
		jwtCfg:   jwtCfg,
	}
}

// Register 注册新用户
func (h *AuthCommandHandler) Register(ctx context.Context, cmd *command.RegisterCommand) (*appRepo.UserDTO, error) {
	// 检查用户名唯一性
	existing, err := h.userRepo.FindByUsername(ctx, cmd.Username)
	if err == nil && existing != nil {
		return nil, ErrUsernameExists
	}

	// 密码加密
	hash, err := crypto.HashPassword(cmd.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	// 创建用户
	user, err := user.NewUser(uuid.New().String(), cmd.Username, hash, cmd.RealName)
	if err != nil {
		return nil, err
	}

	// 分配默认角色（通过读仓储获取默认角色 ID，此为认证流程中的跨聚合查询特例）
	if h.jwtCfg.DefaultRole != "" {
		role, err := h.roleRead.FindByCode(ctx, h.jwtCfg.DefaultRole)
		if err == nil && role != nil {
			user.AssignRole(role.ID)
		}
	}

	if err := h.userRepo.Save(ctx, user); err != nil {
		return nil, err
	}

	return &appRepo.UserDTO{
		ID:        user.ID,
		Username:  user.Username,
		RealName:  user.RealName,
		Status:    user.Status,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

// Login 登录
func (h *AuthCommandHandler) Login(ctx context.Context, cmd *command.LoginCommand) (*LoginResult, error) {
	// 查找用户（含角色权限）
	userWithRoles, err := h.userRead.FindByUsername(ctx, cmd.Username)
	if err != nil {
		// 统一返回"用户名或密码错误"，防止用户枚举
		return nil, appErrors.ErrInvalidCredentials
	}

	// 加载完整用户模型（含 password_hash）
	user, err := h.userRepo.FindByID(ctx, userWithRoles.ID)
	if err != nil {
		return nil, appErrors.ErrInvalidCredentials
	}

	// 验证密码
	if !crypto.VerifyPassword(user.PasswordHash, cmd.Password) {
		return nil, appErrors.ErrInvalidCredentials
	}

	// 检查用户状态
	if !user.IsEnabled() {
		return nil, appErrors.ErrUserDisabled
	}

	// 记录登录时间
	user.RecordLogin()
	_ = h.userRepo.Save(ctx, user) // 保存失败不影响登录

	// 组装 Claims
	roles := make([]string, 0, len(userWithRoles.Roles))
	for _, r := range userWithRoles.Roles {
		roles = append(roles, r.Code)
	}
	permissions := make([]string, 0, len(userWithRoles.Permissions))
	for _, p := range userWithRoles.Permissions {
		permissions = append(permissions, p.Code)
	}

	claims := &auth.Claims{
		UserID:      user.ID,
		Username:    user.Username,
		Roles:       roles,
		Permissions: permissions,
	}

	// 签发令牌对
	tokenPair, err := auth.GenerateTokenPair(
		claims,
		h.jwtCfg.Secret, h.jwtCfg.Expiration,
		h.jwtCfg.Secret, h.jwtCfg.RefreshExpiration,
	)
	if err != nil {
		return nil, fmt.Errorf("generate tokens: %w", err)
	}

	return &LoginResult{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		User: UserInfo{
			ID:          user.ID,
			Username:    user.Username,
			RealName:    user.RealName,
			Roles:       roles,
			Permissions: permissions,
		},
	}, nil
}

// RefreshToken 刷新 AccessToken
func (h *AuthCommandHandler) RefreshToken(ctx context.Context, cmd *command.RefreshTokenCommand) (*auth.TokenPair, error) {
	// 解析 RefreshToken
	claims, err := auth.ParseRefreshToken(cmd.RefreshToken, h.jwtCfg.Secret)
	if err != nil {
		return nil, appErrors.ErrTokenInvalid
	}

	userID := claims.Subject

	// 检查用户是否存在且启用
	userWithRoles, err := h.userRead.FindByIDWithRoles(ctx, userID)
	if err != nil {
		return nil, appErrors.ErrTokenInvalid
	}
	if userWithRoles.Status != user.UserStatusEnabled {
		return nil, appErrors.ErrUserDisabled
	}

	// 重新组装 Claims
	roles := make([]string, 0, len(userWithRoles.Roles))
	for _, r := range userWithRoles.Roles {
		roles = append(roles, r.Code)
	}
	permissions := make([]string, 0, len(userWithRoles.Permissions))
	for _, p := range userWithRoles.Permissions {
		permissions = append(permissions, p.Code)
	}

	accessClaims := &auth.Claims{
		UserID:      userWithRoles.ID,
		Username:    userWithRoles.Username,
		Roles:       roles,
		Permissions: permissions,
	}

	accessToken, err := auth.GenerateAccessToken(accessClaims, h.jwtCfg.Secret, h.jwtCfg.Expiration)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	return &auth.TokenPair{
		AccessToken: accessToken,
		ExpiresIn:   int64(h.jwtCfg.Expiration) * 3600,
	}, nil
}

// Logout 登出（将 AccessToken 加入黑名单）
func (h *AuthCommandHandler) Logout(ctx context.Context, cmd *command.LogoutCommand) error {
	// 解析 AccessToken 获取 jti
	claims, err := auth.ParseToken(cmd.AccessToken, h.jwtCfg.Secret)
	if err != nil {
		// Token 已过期或无效，依然视为登出成功
		return nil
	}

	// 计算剩余 TTL
	now := time.Now()
	remaining := claims.ExpiresAt.Time.Sub(now)
	if remaining <= 0 || h.rdb == nil {
		return nil
	}

	// 将 Token 加入 Redis 黑名单
	key := fmt.Sprintf("token_blacklist:%s", claims.ID)
	return h.rdb.Set(ctx, key, "1", remaining).Err()
}

// LoginResult 登录结果
type LoginResult struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	ExpiresIn    int64    `json:"expires_in"`
	User         UserInfo `json:"user"`
}

// UserInfo 登录返回的用户信息
type UserInfo struct {
	ID          string   `json:"id"`
	Username    string   `json:"username"`
	RealName    string   `json:"real_name"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
}