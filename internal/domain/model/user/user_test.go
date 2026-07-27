package user

import (
	"testing"

	"github.com/google/uuid"
)

func TestNewUser_Success(t *testing.T) {
	user, err := NewUser(uuid.New().String(), "testuser", "hash123", "测试用户")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.Username != "testuser" {
		t.Errorf("expected username testuser, got %s", user.Username)
	}
	if user.RealName != "测试用户" {
		t.Errorf("expected realname 测试用户, got %s", user.RealName)
	}
	if user.Status != UserStatusEnabled {
		t.Errorf("expected status enabled, got %d", user.Status)
	}
	if !user.IsEnabled() {
		t.Error("expected user to be enabled")
	}
	if len(user.UserRoles) != 0 {
		t.Errorf("expected 0 roles, got %d", len(user.UserRoles))
	}
}

func TestNewUser_Validation(t *testing.T) {
	tests := []struct {
		name     string
		username string
		hash     string
		realName string
		wantErr  bool
	}{
		{"empty username", "", "hash", "n", true},
		{"username too long", string(make([]byte, 101)), "hash", "n", true},
		{"empty hash", "user", "", "n", true},
		{"realname too long", "user", "hash", string(make([]byte, 101)), true},
		{"valid", "user", "hash", "n", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewUser(uuid.New().String(), tt.username, tt.hash, tt.realName)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewUser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUser_ChangePassword(t *testing.T) {
	user, _ := NewUser(uuid.New().String(), "user", "oldhash", "n")
	user.ChangePassword("newhash")
	if user.PasswordHash != "newhash" {
		t.Errorf("expected newhash, got %s", user.PasswordHash)
	}
}

func TestUser_UpdateProfile(t *testing.T) {
	user, _ := NewUser(uuid.New().String(), "user", "hash", "n")
	user.UpdateProfile("新名字", "a@b.com", "13800138000")
	if user.RealName != "新名字" || user.Email != "a@b.com" || user.Phone != "13800138000" {
		t.Errorf("profile not updated: %+v", user)
	}
}

func TestUser_RecordLogin(t *testing.T) {
	user, _ := NewUser(uuid.New().String(), "user", "hash", "n")
	user.RecordLogin()
	if user.LastLoginAt == nil {
		t.Error("expected last login at to be set")
	}
}

func TestUser_EnableDisable(t *testing.T) {
	user, _ := NewUser(uuid.New().String(), "user", "hash", "n")
	if !user.IsEnabled() {
		t.Error("expected user to be enabled")
	}
	if err := user.Disable(); err != nil {
		t.Fatalf("disable failed: %v", err)
	}
	if user.IsEnabled() {
		t.Error("expected user to be disabled")
	}
	if err := user.Disable(); err == nil {
		t.Error("expected error on double disable")
	}
	if err := user.Enable(); err != nil {
		t.Fatalf("enable failed: %v", err)
	}
	if !user.IsEnabled() {
		t.Error("expected user to be enabled")
	}
	if err := user.Enable(); err == nil {
		t.Error("expected error on double enable")
	}
}

func TestUser_AssignRole(t *testing.T) {
	user, _ := NewUser(uuid.New().String(), "user", "hash", "n")
	roleID := uuid.New().String()
	user.AssignRole(roleID)
	if len(user.UserRoles) != 1 {
		t.Fatalf("expected 1 role, got %d", len(user.UserRoles))
	}
	if user.UserRoles[0].RoleID != roleID {
		t.Errorf("expected role %s, got %s", roleID, user.UserRoles[0].RoleID)
	}

	// 分配重复角色不应重复添加
	user.AssignRole(roleID)
	if len(user.UserRoles) != 1 {
		t.Errorf("expected 1 role after duplicate assign, got %d", len(user.UserRoles))
	}
}

func TestUser_RemoveRole(t *testing.T) {
	user, _ := NewUser(uuid.New().String(), "user", "hash", "n")
	role1 := uuid.New().String()
	role2 := uuid.New().String()
	user.AssignRole(role1)
	user.AssignRole(role2)
	if len(user.UserRoles) != 2 {
		t.Fatalf("expected 2 roles, got %d", len(user.UserRoles))
	}

	user.RemoveRole(role1)
	if len(user.UserRoles) != 1 {
		t.Errorf("expected 1 role after remove, got %d", len(user.UserRoles))
	}
	if user.UserRoles[0].RoleID != role2 {
		t.Errorf("expected remaining role to be %s, got %s", role2, user.UserRoles[0].RoleID)
	}
}

func TestUser_GetRoleIDs(t *testing.T) {
	user, _ := NewUser(uuid.New().String(), "user", "hash", "n")
	role1 := uuid.New().String()
	role2 := uuid.New().String()
	user.AssignRole(role1)
	user.AssignRole(role2)
	ids := user.GetRoleIDs()
	if len(ids) != 2 {
		t.Fatalf("expected 2 ids, got %d", len(ids))
	}
}