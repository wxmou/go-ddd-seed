package auth

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"time"
)

// Claims 自定义 JWT Claims
type Claims struct {
	UserID      string   `json:"userID"`
	Username    string   `json:"username"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

// TokenPair 令牌对（AccessToken + RefreshToken）
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // AccessToken 过期秒数
}

// GenerateAccessToken 生成 AccessToken
func GenerateAccessToken(claims *Claims, secret string, expHours int) (string, error) {
	now := time.Now()
	claims.IssuedAt = jwt.NewNumericDate(now)
	claims.ExpiresAt = jwt.NewNumericDate(now.Add(time.Duration(expHours) * time.Hour))
	claims.ID = uuid.New().String() // jti

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// GenerateRefreshToken 生成 RefreshToken
// RefreshToken 仅包含 userID，不包含角色权限信息，有效期更长
func GenerateRefreshToken(userID string, secret string, expHours int) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   userID,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(expHours) * time.Hour)),
		ID:        uuid.New().String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ParseToken 解析 JWT Token（通用）
func ParseToken(tokenString string, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}
	return claims, nil
}

// ParseRefreshToken 解析 RefreshToken（返回 RegisteredClaims）
func ParseRefreshToken(tokenString string, secret string) (*jwt.RegisteredClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}
	return claims, nil
}

// GenerateTokenPair 生成完整的 AccessToken + RefreshToken 对
func GenerateTokenPair(claims *Claims, accessSecret string, accessExpHours int, refreshSecret string, refreshExpHours int) (*TokenPair, error) {
	accessToken, err := GenerateAccessToken(claims, accessSecret, accessExpHours)
	if err != nil {
		return nil, err
	}

	refreshToken, err := GenerateRefreshToken(claims.UserID, refreshSecret, refreshExpHours)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(accessExpHours) * 3600,
	}, nil
}

// GetTokenID 从 Claims 中提取 jti
func GetTokenID(claims *Claims) string {
	return claims.ID
}