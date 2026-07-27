package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-ddd-seed/go-ddd-seed/internal/application/command"
	"github.com/go-ddd-seed/go-ddd-seed/internal/application/commandHandler"
	"github.com/go-ddd-seed/go-ddd-seed/internal/trigger/http/req"
	"github.com/go-ddd-seed/go-ddd-seed/internal/trigger/http/resp"
	"github.com/go-ddd-seed/go-ddd-seed/pkg/middleware"
	"github.com/go-ddd-seed/go-ddd-seed/pkg/utils"
	"github.com/go-ddd-seed/go-ddd-seed/pkg/validator"
)

// AuthController 认证控制器
type AuthController struct {
	handler *commandHandler.AuthCommandHandler
}

// NewAuthController 创建认证控制器
func NewAuthController(handler *commandHandler.AuthCommandHandler) *AuthController {
	return &AuthController{handler: handler}
}

// RegisterRoutes 注册路由
func (c *AuthController) RegisterRoutes(public, auth *gin.RouterGroup) {
	// 公开路由（无需认证）
	public.POST("/auth/register", c.Register)
	public.POST("/auth/login", c.Login)
	public.POST("/auth/refresh", c.RefreshToken)

	// 需认证路由
	auth.POST("/auth/logout", c.Logout)
	auth.GET("/auth/me", c.Me)
}

// Register 注册
// @Summary      用户注册
// @Description  创建新的用户账号
// @Tags         认证管理
// @Accept       json
// @Produce      json
// @Param        request  body  req.RegisterReq  true  "注册请求参数"
// @Success      200  {object}  resp.UserInfo   "注册成功，返回用户信息"
// @Failure      400  {object}  docs.APIError   "请求参数错误"
// @Failure      409  {object}  docs.APIError   "用户名已存在"
// @Router       /auth/register [post]
func (c *AuthController) Register(ctx *gin.Context) {
	var registerReq req.RegisterReq
	if err := ctx.ShouldBindJSON(&registerReq); err != nil {
		utils.FailWithMsg(ctx, http.StatusBadRequest, 2, err.Error())
		return
	}
	if err := validator.ValidateRequest(registerReq); err != nil {
		utils.ValidationFail(ctx, err)
		return
	}

	cmd := &command.RegisterCommand{
		Username: registerReq.Username,
		Password: registerReq.Password,
		RealName: registerReq.RealName,
	}

	result, err := c.handler.Register(ctx.Request.Context(), cmd)
	if err != nil {
		utils.Error(ctx, err)
		return
	}
	utils.Success(ctx, result)
}

// Login 登录
// @Summary      用户登录
// @Description  使用用户名和密码登录，返回 JWT 访问令牌和刷新令牌
// @Tags         认证管理
// @Accept       json
// @Produce      json
// @Param        request  body  req.LoginReq  true  "登录请求参数"
// @Success      200  {object}  resp.LoginResp  "登录成功"
// @Failure      400  {object}  docs.APIError   "请求参数错误"
// @Failure      401  {object}  docs.APIError   "用户名或密码错误"
// @Router       /auth/login [post]
func (c *AuthController) Login(ctx *gin.Context) {
	var loginReq req.LoginReq
	if err := ctx.ShouldBindJSON(&loginReq); err != nil {
		utils.FailWithMsg(ctx, http.StatusBadRequest, 2, err.Error())
		return
	}
	if err := validator.ValidateRequest(loginReq); err != nil {
		utils.ValidationFail(ctx, err)
		return
	}

	cmd := &command.LoginCommand{
		Username: loginReq.Username,
		Password: loginReq.Password,
	}

	result, err := c.handler.Login(ctx.Request.Context(), cmd)
	if err != nil {
		utils.Error(ctx, err)
		return
	}
	utils.Success(ctx, toLoginResp(result))
}

// RefreshToken 刷新令牌
// @Summary      刷新访问令牌
// @Description  使用刷新令牌获取新的访问令牌
// @Tags         认证管理
// @Accept       json
// @Produce      json
// @Param        request  body  req.RefreshTokenReq  true  "刷新令牌请求参数"
// @Success      200  {object}  resp.TokenResp  "刷新成功，返回新的访问令牌"
// @Failure      400  {object}  docs.APIError  "请求参数错误"
// @Failure      401  {object}  docs.APIError  "刷新令牌无效或已过期"
// @Router       /auth/refresh [post]
func (c *AuthController) RefreshToken(ctx *gin.Context) {
	var refreshReq req.RefreshTokenReq
	if err := ctx.ShouldBindJSON(&refreshReq); err != nil {
		utils.FailWithMsg(ctx, http.StatusBadRequest, 2, err.Error())
		return
	}
	if err := validator.ValidateRequest(refreshReq); err != nil {
		utils.ValidationFail(ctx, err)
		return
	}

	cmd := &command.RefreshTokenCommand{
		RefreshToken: refreshReq.RefreshToken,
	}

	tokenPair, err := c.handler.RefreshToken(ctx.Request.Context(), cmd)
	if err != nil {
		utils.Error(ctx, err)
		return
	}
	utils.Success(ctx, toTokenResp(tokenPair.AccessToken, tokenPair.ExpiresIn))
}

// Logout 登出
// @Summary      退出登录
// @Description  撤销当前访问令牌，退出登录状态
// @Tags         认证管理
// @Accept       json
// @Produce      json
// @Success      200  {object}  docs.APIError  "登出成功"
// @Security     ApiKeyAuth
// @Router       /auth/logout [post]
func (c *AuthController) Logout(ctx *gin.Context) {
	// 从请求头提取 AccessToken
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		utils.Success(ctx, nil) // 视为已登出
		return
	}

	// 从上下文中获取用户信息（由中间件注入）
	_, ok := middleware.GetCurrentUserID(ctx)
	if !ok {
		utils.Success(ctx, nil)
		return
	}

	tokenStr := authHeader[len("Bearer "):]
	cmd := &command.LogoutCommand{
		AccessToken: tokenStr,
	}

	if err := c.handler.Logout(ctx.Request.Context(), cmd); err != nil {
		utils.Error(ctx, err)
		return
	}
	utils.Success(ctx, nil)
}

// Me 获取当前用户信息
// @Summary      获取当前用户信息
// @Description  获取已登录用户的详细信息，包括角色和权限
// @Tags         认证管理
// @Produce      json
// @Success      200  {object}  resp.UserInfo  "当前用户信息"
// @Failure      401  {object}  docs.APIError  "未授权"
// @Security     ApiKeyAuth
// @Router       /auth/me [get]
func (c *AuthController) Me(ctx *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(ctx)
	if !ok {
		utils.FailWithMsg(ctx, http.StatusUnauthorized, 3, "未授权")
		return
	}

	username, _ := middleware.GetCurrentUsername(ctx)
	roles, _ := middleware.GetCurrentRoles(ctx)
	permissions, _ := middleware.GetCurrentPermissions(ctx)

	utils.Success(ctx, &resp.UserInfo{
		ID:          userID,
		Username:    username,
		Roles:       roles,
		Permissions: permissions,
	})
}

// toLoginResp 将应用层 LoginResult 转换为 HTTP 响应
func toLoginResp(result *commandHandler.LoginResult) *resp.LoginResp {
	return &resp.LoginResp{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresIn:    result.ExpiresIn,
		User: resp.UserInfo{
			ID:          result.User.ID,
			Username:    result.User.Username,
			RealName:    result.User.RealName,
			Roles:       result.User.Roles,
			Permissions: result.User.Permissions,
		},
	}
}

// toTokenResp 令牌对转响应
func toTokenResp(accessToken string, expiresIn int64) *resp.TokenResp {
	return &resp.TokenResp{
		AccessToken: accessToken,
		ExpiresIn:   expiresIn,
	}
}