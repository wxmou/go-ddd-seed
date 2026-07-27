package middleware

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	appErrors "github.com/go-ddd-seed/go-ddd-seed/pkg/errors"
	"net/http"

	pkgAuth "github.com/go-ddd-seed/go-ddd-seed/pkg/auth"
)

// Context keys for user info injected by auth middleware
const (
	KeyUserID      = "userID"
	KeyUsername    = "username"
	KeyRoles       = "roles"
	KeyPermissions = "permissions"
)

// Auth JWT 认证中间件（增强版：支持 Redis 黑名单检查 + 完整 Claims 注入）
func Auth(secret string, rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			abortUnauthorized(c, appErrors.ErrUnauthorized.Code, appErrors.ErrUnauthorized.Message)
			return
		}

		// Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			abortUnauthorized(c, appErrors.ErrTokenInvalid.Code, appErrors.ErrTokenInvalid.Message)
			return
		}

		tokenString := parts[1]

		// 使用 pkg/auth 解析自定义 Claims
		claims, err := pkgAuth.ParseToken(tokenString, secret)
		if err != nil {
			code := appErrors.ErrTokenInvalid.Code
			msg := appErrors.ErrTokenInvalid.Message
			if strings.Contains(err.Error(), "expired") {
				code = appErrors.ErrTokenExpired.Code
				msg = appErrors.ErrTokenExpired.Message
			}
			abortUnauthorized(c, code, msg)
			return
		}

		// 检查 Redis 黑名单（Token 是否被撤销）
		if rdb != nil && claims.ID != "" {
			key := fmt.Sprintf("token_blacklist:%s", claims.ID)
			exists, ceErr := rdb.Exists(c.Request.Context(), key).Result()
			if ceErr == nil && exists > 0 {
				abortUnauthorized(c, appErrors.ErrTokenExpired.Code, "令牌已被撤销")
				return
			}
		}

		// 注入用户信息到 Context
		c.Set(KeyUserID, claims.UserID)
		c.Set(KeyUsername, claims.Username)
		c.Set(KeyRoles, claims.Roles)
		c.Set(KeyPermissions, claims.Permissions)

		c.Next()
	}
}

func abortUnauthorized(c *gin.Context, code int, message string) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"code":    code,
		"message": message,
	})
}

// GetCurrentUserID 从 Context 获取当前用户 ID
func GetCurrentUserID(c *gin.Context) (string, bool) {
	v, exists := c.Get(KeyUserID)
	if !exists {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

// GetCurrentUsername 从 Context 获取当前用户名
func GetCurrentUsername(c *gin.Context) (string, bool) {
	v, exists := c.Get(KeyUsername)
	if !exists {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

// GetCurrentRoles 从 Context 获取当前用户角色
func GetCurrentRoles(c *gin.Context) ([]string, bool) {
	v, exists := c.Get(KeyRoles)
	if !exists {
		return nil, false
	}
	s, ok := v.([]string)
	return s, ok
}

// GetCurrentPermissions 从 Context 获取当前用户权限
func GetCurrentPermissions(c *gin.Context) ([]string, bool) {
	v, exists := c.Get(KeyPermissions)
	if !exists {
		return nil, false
	}
	s, ok := v.([]string)
	return s, ok
}