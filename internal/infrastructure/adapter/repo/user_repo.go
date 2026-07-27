package repo

import (
	"context"
	"time"

	"github.com/go-ddd-seed/go-ddd-seed/internal/domain"
	"gorm.io/gorm"

	domainRepo "github.com/go-ddd-seed/go-ddd-seed/internal/domain/port/repo"
	"github.com/go-ddd-seed/go-ddd-seed/internal/domain/model/user"
	
)

// Ensure interface compliance
var _ domainRepo.UserRepository = (*UserRepository)(nil)

// UserGorm 用户 GORM 模型
type UserGorm struct {
	ID           string     `gorm:"column:id;primaryKey;type:varchar(36)"`
	Username     string     `gorm:"column:username;type:varchar(100);uniqueIndex;not null"`
	PasswordHash string     `gorm:"column:password_hash;type:varchar(255);not null"`
	RealName     string     `gorm:"column:real_name;type:varchar(100)"`
	Email        string     `gorm:"column:email;type:varchar(200)"`
	Phone        string     `gorm:"column:phone;type:varchar(50)"`
	Status       int        `gorm:"column:status;default:1"`
	LastLoginAt  *time.Time `gorm:"column:last_login_at"`
	CreatedAt    time.Time  `gorm:"column:created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at"`
}

// TableName 表名
func (UserGorm) TableName() string {
	return "users"
}

// UserRoleGorm 用户-角色关联 GORM 模型
type UserRoleGorm struct {
	ID        string    `gorm:"column:id;primaryKey;type:varchar(36)"`
	UserID    string    `gorm:"column:user_id;type:varchar(36);uniqueIndex:uk_user_role;not null"`
	RoleID    string    `gorm:"column:role_id;type:varchar(36);uniqueIndex:uk_user_role;not null"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

// TableName 表名
func (UserRoleGorm) TableName() string {
	return "user_roles"
}

func fromUserDomain(u *user.User) *UserGorm {
	return &UserGorm{
		ID:           u.ID,
		Username:     u.Username,
		PasswordHash: u.PasswordHash,
		RealName:     u.RealName,
		Email:        u.Email,
		Phone:        u.Phone,
		Status:       u.Status,
		LastLoginAt:  u.LastLoginAt,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
}

func fromUserRoleDomain(ur *user.UserRole) *UserRoleGorm {
	return &UserRoleGorm{
		ID:        ur.ID,
		UserID:    ur.UserID,
		RoleID:    ur.RoleID,
		CreatedAt: ur.CreatedAt,
	}
}

// UserRepository GORM 仓储实现（命令侧）
type UserRepository struct {
	RepositoryBase
}

// NewUserRepository 创建仓储
func NewUserRepository(base RepositoryBase) *UserRepository {
	return &UserRepository{RepositoryBase: base}
}

// FindByID 按 ID 加载用户（含角色关联）
func (r *UserRepository) FindByID(ctx context.Context, id string) (*user.User, error) {
	var gormUser UserGorm
	if err := r.DB.WithContext(ctx).Where("id = ?", id).First(&gormUser).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrRecordNotFound
		}
		return nil, err
	}

	// 加载角色关联
	var gormRoles []UserRoleGorm
	if err := r.DB.WithContext(ctx).Where("user_id = ?", id).Find(&gormRoles).Error; err != nil {
		return nil, err
	}

	userRoles := make([]*user.UserRole, 0, len(gormRoles))
	for _, gr := range gormRoles {
		userRoles = append(userRoles, &user.UserRole{
			ID:        gr.ID,
			UserID:    gr.UserID,
			RoleID:    gr.RoleID,
			CreatedAt: gr.CreatedAt,
		})
	}

	return &user.User{
		ID:           gormUser.ID,
		Username:     gormUser.Username,
		PasswordHash: gormUser.PasswordHash,
		RealName:     gormUser.RealName,
		Email:        gormUser.Email,
		Phone:        gormUser.Phone,
		Status:       gormUser.Status,
		LastLoginAt:  gormUser.LastLoginAt,
		UserRoles:    userRoles,
		CreatedAt:    gormUser.CreatedAt,
		UpdatedAt:    gormUser.UpdatedAt,
	}, nil
}

// FindByUsername 按用户名加载用户（含角色关联）
func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*user.User, error) {
	var gormUser UserGorm
	if err := r.DB.WithContext(ctx).Where("username = ?", username).First(&gormUser).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrRecordNotFound
		}
		return nil, err
	}

	// 加载角色关联
	var gormRoles []UserRoleGorm
	if err := r.DB.WithContext(ctx).Where("user_id = ?", gormUser.ID).Find(&gormRoles).Error; err != nil {
		return nil, err
	}

	userRoles := make([]*user.UserRole, 0, len(gormRoles))
	for _, gr := range gormRoles {
		userRoles = append(userRoles, &user.UserRole{
			ID:        gr.ID,
			UserID:    gr.UserID,
			RoleID:    gr.RoleID,
			CreatedAt: gr.CreatedAt,
		})
	}

	return &user.User{
		ID:           gormUser.ID,
		Username:     gormUser.Username,
		PasswordHash: gormUser.PasswordHash,
		RealName:     gormUser.RealName,
		Email:        gormUser.Email,
		Phone:        gormUser.Phone,
		Status:       gormUser.Status,
		LastLoginAt:  gormUser.LastLoginAt,
		UserRoles:    userRoles,
		CreatedAt:    gormUser.CreatedAt,
		UpdatedAt:    gormUser.UpdatedAt,
	}, nil
}

// Save 保存用户（含角色关联，整存整取），自动发布领域事件
func (r *UserRepository) Save(ctx context.Context, user *user.User) error {
	return r.SaveWithEvents(ctx, user, func(tx *gorm.DB) error {
		// 保存用户基本信息
		gormUser := fromUserDomain(user)
		if err := tx.Save(gormUser).Error; err != nil {
			return err
		}

		// 全量替换角色关联
		if err := tx.Where("user_id = ?", user.ID).Delete(&UserRoleGorm{}).Error; err != nil {
			return err
		}

		for _, ur := range user.UserRoles {
			gormRole := fromUserRoleDomain(ur)
			gormRole.UserID = user.ID
			if err := tx.Create(gormRole).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// Delete 删除用户（级联删除角色关联）
func (r *UserRepository) Delete(ctx context.Context, id string) error {
	return r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 先删除角色关联
		if err := tx.Where("user_id = ?", id).Delete(&UserRoleGorm{}).Error; err != nil {
			return err
		}

		result := tx.Where("id = ?", id).Delete(&UserGorm{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return domain.ErrRecordNotFound
		}
		return nil
	})
}