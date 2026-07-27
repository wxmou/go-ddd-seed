package repo

import (
	"context"

	"github.com/go-ddd-seed/go-ddd-seed/internal/domain/model/user"
)

// UserRepository 用户仓储接口（命令侧）
type UserRepository interface {
	// Save 保存用户（含角色关联）
	Save(ctx context.Context, user *user.User) error
	// FindByID 按 ID 加载用户（含角色关联）
	FindByID(ctx context.Context, id string) (*user.User, error)
	// FindByUsername 按用户名加载用户
	FindByUsername(ctx context.Context, username string) (*user.User, error)
	// Delete 删除用户（级联删除角色关联）
	Delete(ctx context.Context, id string) error
}