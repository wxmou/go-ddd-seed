package user

import (
	"time"

	"github.com/google/uuid"
	"github.com/go-ddd-seed/go-ddd-seed/internal/domain/model"
)

// UserStatus 用户状态
const (
	UserStatusDisabled = 0
	UserStatusEnabled  = 1
)

// User 用户聚合根
// 包含 UserRole 子实体列表
type User struct {
	model.AggregateRoot
	ID           string
	Username     string
	PasswordHash string
	RealName     string
	Email        string
	Phone        string
	Status       int
	LastLoginAt  *time.Time
	UserRoles    []*UserRole
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// NewUser 创建新用户
func NewUser(id, username, passwordHash, realName string) (*User, error) {
	if username == "" {
		return nil, ErrUsernameEmpty
	}
	if len(username) > 100 {
		return nil, ErrUsernameTooLong
	}
	if len(passwordHash) == 0 {
		return nil, ErrPasswordTooShort
	}
	if len(realName) > 100 {
		return nil, ErrRealNameTooLong
	}

	now := time.Now()
	return &User{
		ID:           id,
		Username:     username,
		PasswordHash: passwordHash,
		RealName:     realName,
		Status:       UserStatusEnabled,
		UserRoles:    make([]*UserRole, 0),
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// ChangePassword 修改密码
func (u *User) ChangePassword(newHash string) {
	u.PasswordHash = newHash
	u.UpdatedAt = time.Now()
}

// UpdateProfile 更新个人信息
func (u *User) UpdateProfile(realName, email, phone string) {
	u.RealName = realName
	u.Email = email
	u.Phone = phone
	u.UpdatedAt = time.Now()
}

// RecordLogin 记录登录时间
func (u *User) RecordLogin() {
	now := time.Now()
	u.LastLoginAt = &now
	u.UpdatedAt = now
}

// Enable 启用用户
func (u *User) Enable() error {
	if u.Status == UserStatusEnabled {
		return ErrUserAlreadyEnabled
	}
	u.Status = UserStatusEnabled
	u.UpdatedAt = time.Now()
	return nil
}

// Disable 禁用用户
func (u *User) Disable() error {
	if u.Status == UserStatusDisabled {
		return ErrUserAlreadyDisabled
	}
	u.Status = UserStatusDisabled
	u.UpdatedAt = time.Now()
	return nil
}

// IsEnabled 是否启用
func (u *User) IsEnabled() bool {
	return u.Status == UserStatusEnabled
}

// AssignRole 分配角色
func (u *User) AssignRole(roleID string) {
	// 检查是否已分配
	for _, ur := range u.UserRoles {
		if ur.RoleID == roleID {
			return
		}
	}
	u.UserRoles = append(u.UserRoles, &UserRole{
		ID:        uuid.New().String(),
		UserID:    u.ID,
		RoleID:    roleID,
		CreatedAt: time.Now(),
	})
	u.UpdatedAt = time.Now()
}

// RemoveRole 移除角色
func (u *User) RemoveRole(roleID string) {
	for i, ur := range u.UserRoles {
		if ur.RoleID == roleID {
			u.UserRoles = append(u.UserRoles[:i], u.UserRoles[i+1:]...)
			u.UpdatedAt = time.Now()
			return
		}
	}
}

// GetRoleIDs 获取所有角色 ID
func (u *User) GetRoleIDs() []string {
	ids := make([]string, 0, len(u.UserRoles))
	for _, ur := range u.UserRoles {
		ids = append(ids, ur.RoleID)
	}
	return ids
}