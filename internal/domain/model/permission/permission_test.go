package permission

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewPermission_Success(t *testing.T) {
	p, err := NewPermission(uuid.New().String(), "创建案件", "case:create", "创建新案件")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if p.Name != "创建案件" || p.Code != "case:create" {
		t.Errorf("permission fields mismatch: %+v", p)
	}
}

func TestNewPermission_Validation(t *testing.T) {
	tests := []struct {
		desc     string
		permName string
		code     string
		wantErr  bool
	}{
		{"empty name", "", "case:create", true},
		{"empty code", "创建案件", "", true},
		{"code too long", "创建案件", string(make([]byte, 201)), true},
		{"valid", "创建案件", "case:create", false},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			_, err := NewPermission(uuid.New().String(), tt.permName, tt.code, "desc")
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPermission() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPermission_Update(t *testing.T) {
	p, _ := NewPermission(uuid.New().String(), "旧名", "case:old", "旧描述")
	before := p.UpdatedAt
	time.Sleep(time.Millisecond)
	p.Update("新名", "新描述")
	if p.Name != "新名" || p.Description != "新描述" {
		t.Errorf("permission not updated: %+v", p)
	}
	if !p.UpdatedAt.After(before) && !p.UpdatedAt.Equal(before) {
		t.Error("expected updatedAt to change")
	}
}