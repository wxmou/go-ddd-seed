package kv_config

import (
	"testing"
)

func TestKvConfig_NewKvConfig(t *testing.T) {
	t.Run("success with valid key and value", func(t *testing.T) {
		kv, err := NewKvConfig("id-1", "site_name", "TestApp", "站点名称")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if kv.ID != "id-1" {
			t.Errorf("expected id 'id-1', got %q", kv.ID)
		}
		if kv.Key != "site_name" {
			t.Errorf("expected key 'site_name', got %q", kv.Key)
		}
		if kv.Value != "TestApp" {
			t.Errorf("expected value 'TestApp', got %q", kv.Value)
		}
		if kv.Description != "站点名称" {
			t.Errorf("expected description '站点名称', got %q", kv.Description)
		}
		if kv.Status != KvConfigStatusEnabled {
			t.Errorf("expected status %d (enabled), got %d", KvConfigStatusEnabled, kv.Status)
		}
		if kv.CreatedAt.IsZero() {
			t.Error("expected non-zero CreatedAt")
		}
	})

	t.Run("error when key is empty", func(t *testing.T) {
		_, err := NewKvConfig("id-2", "", "value", "")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err != ErrKvConfigKeyEmpty {
			t.Errorf("expected ErrKvConfigKeyEmpty, got %v", err)
		}
	})

	t.Run("empty value is allowed", func(t *testing.T) {
		kv, err := NewKvConfig("id-3", "empty_key", "", "")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if kv.Value != "" {
			t.Errorf("expected empty value, got %q", kv.Value)
		}
	})
}

func TestKvConfig_Update(t *testing.T) {
	kv, _ := NewKvConfig("id-1", "site_name", "TestApp", "")
	oldUpdatedAt := kv.UpdatedAt

	kv.Update("NewName", "新描述")

	if kv.Value != "NewName" {
		t.Errorf("expected value 'NewName', got %q", kv.Value)
	}
	if kv.Description != "新描述" {
		t.Errorf("expected description '新描述', got %q", kv.Description)
	}
	if !kv.UpdatedAt.After(oldUpdatedAt) && !kv.UpdatedAt.Equal(oldUpdatedAt) {
		t.Error("expected UpdatedAt to be updated")
	}
}

func TestKvConfig_EnableDisable(t *testing.T) {
	t.Run("enable on already enabled returns error", func(t *testing.T) {
		kv, _ := NewKvConfig("id-1", "site_name", "TestApp", "")
		err := kv.Enable()
		if err != ErrKvConfigAlreadyEnabled {
			t.Errorf("expected ErrKvConfigAlreadyEnabled, got %v", err)
		}
	})

	t.Run("disable then enable", func(t *testing.T) {
		kv, _ := NewKvConfig("id-2", "site_name", "TestApp", "")

		err := kv.Disable()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if kv.Status != KvConfigStatusDisabled {
			t.Errorf("expected status disabled, got %d", kv.Status)
		}
		if kv.IsEnabled() {
			t.Error("expected IsEnabled to be false")
		}

		err = kv.Enable()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if kv.Status != KvConfigStatusEnabled {
			t.Errorf("expected status enabled, got %d", kv.Status)
		}
		if !kv.IsEnabled() {
			t.Error("expected IsEnabled to be true")
		}
	})

	t.Run("disable on already disabled returns error", func(t *testing.T) {
		kv, _ := NewKvConfig("id-3", "site_name", "TestApp", "")
		_ = kv.Disable()

		err := kv.Disable()
		if err != ErrKvConfigAlreadyDisabled {
			t.Errorf("expected ErrKvConfigAlreadyDisabled, got %v", err)
		}
	})
}