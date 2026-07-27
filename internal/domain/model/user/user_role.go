package user

import "time"

// UserRole 用户-角色关联实体
type UserRole struct {
	ID        string
	UserID    string
	RoleID    string
	CreatedAt time.Time
}