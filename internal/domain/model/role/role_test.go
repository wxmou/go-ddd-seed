package role

import (
	"testing"

	"github.com/google/uuid"
)

func TestNewRole_Success(t *testing.T) {
	role, err := NewRole(uuid.New().String(), "管理员", "admin", "系统管理员")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if role.Name != "管理员" || role.Code != "admin" {
		t.Errorf("role fields mismatch: %+v", role)
	}
	if !role.IsEnabled() {
		t.Error("expected role to be enabled")
	}
}

func TestNewRole_Validation(t *testing.T) {
	tests := []struct {
		desc     string
		roleName string
		code     string
		wantErr  bool
	}{
		{"empty name", "", "admin", true},
		{"empty code", "admin", "", true},
		{"code too long", "admin", string(make([]byte, 101)), true},
		{"valid", "admin", "user", false},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			_, err := NewRole(uuid.New().String(), tt.roleName, tt.code, "desc")
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRole() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRole_AssignPermission(t *testing.T) {
	role, _ := NewRole(uuid.New().String(), "admin", "admin", "")
	permID := uuid.New().String()
	role.AssignPermission(permID)
	if len(role.Permissions) != 1 {
		t.Fatalf("expected 1 permission, got %d", len(role.Permissions))
	}
	role.AssignPermission(permID)
	if len(role.Permissions) != 1 {
		t.Errorf("expected 1 after duplicate, got %d", len(role.Permissions))
	}
}

func TestRole_RemovePermission(t *testing.T) {
	role, _ := NewRole(uuid.New().String(), "admin", "admin", "")
	perm1 := uuid.New().String()
	perm2 := uuid.New().String()
	role.AssignPermission(perm1)
	role.AssignPermission(perm2)
	role.RemovePermission(perm1)
	if len(role.Permissions) != 1 {
		t.Errorf("expected 1 after remove, got %d", len(role.Permissions))
	}
}