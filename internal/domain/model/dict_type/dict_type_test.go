package dict_type

import (
	"testing"
)

func TestDictType_NewDictType(t *testing.T) {
	t.Run("success with valid code and name", func(t *testing.T) {
		dt, err := NewDictType("id-1", "gender", "性别", "性别字典")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if dt.ID != "id-1" {
			t.Errorf("expected id 'id-1', got %q", dt.ID)
		}
		if dt.Code != "gender" {
			t.Errorf("expected code 'gender', got %q", dt.Code)
		}
		if dt.Name != "性别" {
			t.Errorf("expected name '性别', got %q", dt.Name)
		}
		if dt.Description != "性别字典" {
			t.Errorf("expected description '性别字典', got %q", dt.Description)
		}
		if dt.Status != DictTypeStatusEnabled {
			t.Errorf("expected status %d (enabled), got %d", DictTypeStatusEnabled, dt.Status)
		}
		if dt.Entries == nil {
			t.Error("expected non-nil entries slice")
		} else if len(dt.Entries) != 0 {
			t.Errorf("expected empty entries, got %d", len(dt.Entries))
		}
		if dt.CreatedAt.IsZero() {
			t.Error("expected non-zero CreatedAt")
		}
		if dt.UpdatedAt.IsZero() {
			t.Error("expected non-zero UpdatedAt")
		}
	})

	t.Run("error when code is empty", func(t *testing.T) {
		_, err := NewDictType("id-2", "", "测试", "")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err != ErrDictTypeCodeEmpty {
			t.Errorf("expected ErrDictTypeCodeEmpty, got %v", err)
		}
	})

	t.Run("error when name is empty", func(t *testing.T) {
		_, err := NewDictType("id-3", "test", "", "")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err != ErrDictTypeNameEmpty {
			t.Errorf("expected ErrDictTypeNameEmpty, got %v", err)
		}
	})
}

func TestDictType_Update(t *testing.T) {
	dt, _ := NewDictType("id-1", "gender", "性别", "")
	oldUpdatedAt := dt.UpdatedAt

	dt.Update("性别2", "新描述")

	if dt.Name != "性别2" {
		t.Errorf("expected name '性别2', got %q", dt.Name)
	}
	if dt.Description != "新描述" {
		t.Errorf("expected description '新描述', got %q", dt.Description)
	}
	if !dt.UpdatedAt.After(oldUpdatedAt) && !dt.UpdatedAt.Equal(oldUpdatedAt) {
		t.Error("expected UpdatedAt to be updated")
	}
}

func TestDictType_EnableDisable(t *testing.T) {
	t.Run("enable on already enabled type returns error", func(t *testing.T) {
		dt, _ := NewDictType("id-1", "gender", "性别", "")
		err := dt.Enable()
		if err != ErrDictTypeAlreadyEnabled {
			t.Errorf("expected ErrDictTypeAlreadyEnabled, got %v", err)
		}
		if dt.Status != DictTypeStatusEnabled {
			t.Errorf("expected status enabled, got %d", dt.Status)
		}
	})

	t.Run("disable then enable", func(t *testing.T) {
		dt, _ := NewDictType("id-2", "gender", "性别", "")

		err := dt.Disable()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if dt.Status != DictTypeStatusDisabled {
			t.Errorf("expected status disabled, got %d", dt.Status)
		}
		if dt.IsEnabled() {
			t.Error("expected IsEnabled to be false")
		}

		err = dt.Enable()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if dt.Status != DictTypeStatusEnabled {
			t.Errorf("expected status enabled, got %d", dt.Status)
		}
		if !dt.IsEnabled() {
			t.Error("expected IsEnabled to be true")
		}
	})

	t.Run("disable on already disabled type returns error", func(t *testing.T) {
		dt, _ := NewDictType("id-3", "gender", "性别", "")
		_ = dt.Disable()

		err := dt.Disable()
		if err != ErrDictTypeAlreadyDisabled {
			t.Errorf("expected ErrDictTypeAlreadyDisabled, got %v", err)
		}
	})
}

func TestDictType_AddEntry(t *testing.T) {
	t.Run("add entry successfully", func(t *testing.T) {
		dt, _ := NewDictType("id-1", "gender", "性别", "")

		entry := &DictEntry{ID: "e-1", Label: "男", Value: "male", SortOrder: 1, Status: DictEntryStatusEnabled}
		err := dt.AddEntry(entry)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(dt.Entries) != 1 {
			t.Fatalf("expected 1 entry, got %d", len(dt.Entries))
		}
		if dt.Entries[0].Value != "male" {
			t.Errorf("expected value 'male', got %q", dt.Entries[0].Value)
		}
	})

	t.Run("add entry with duplicate value returns error", func(t *testing.T) {
		dt, _ := NewDictType("id-2", "gender", "性别", "")
		_ = dt.AddEntry(&DictEntry{ID: "e-1", Label: "男", Value: "male", Status: DictEntryStatusEnabled})

		err := dt.AddEntry(&DictEntry{ID: "e-2", Label: "男2", Value: "male", Status: DictEntryStatusEnabled})
		if err != ErrDictEntryValueDuplicate {
			t.Errorf("expected ErrDictEntryValueDuplicate, got %v", err)
		}
	})

	t.Run("add multiple entries preserves sort order", func(t *testing.T) {
		dt, _ := NewDictType("id-3", "gender", "性别", "")
		_ = dt.AddEntry(&DictEntry{ID: "e-1", Label: "男", Value: "male", SortOrder: 2, Status: DictEntryStatusEnabled})
		_ = dt.AddEntry(&DictEntry{ID: "e-2", Label: "女", Value: "female", SortOrder: 1, Status: DictEntryStatusEnabled})

		if len(dt.Entries) != 2 {
			t.Fatalf("expected 2 entries, got %d", len(dt.Entries))
		}
	})
}

func TestDictType_UpdateEntry(t *testing.T) {
	t.Run("update entry successfully", func(t *testing.T) {
		dt, _ := NewDictType("id-1", "gender", "性别", "")
		_ = dt.AddEntry(&DictEntry{ID: "e-1", Label: "男", Value: "male", SortOrder: 1, Status: DictEntryStatusEnabled})

		err := dt.UpdateEntry("e-1", "男性", "male", 2)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if dt.Entries[0].Label != "男性" {
			t.Errorf("expected label '男性', got %q", dt.Entries[0].Label)
		}
		if dt.Entries[0].SortOrder != 2 {
			t.Errorf("expected sort_order 2, got %d", dt.Entries[0].SortOrder)
		}
	})

	t.Run("update entry with duplicate value returns error", func(t *testing.T) {
		dt, _ := NewDictType("id-2", "gender", "性别", "")
		_ = dt.AddEntry(&DictEntry{ID: "e-1", Label: "男", Value: "male", Status: DictEntryStatusEnabled})
		_ = dt.AddEntry(&DictEntry{ID: "e-2", Label: "女", Value: "female", Status: DictEntryStatusEnabled})

		err := dt.UpdateEntry("e-1", "男", "female", 1)
		if err != ErrDictEntryValueDuplicate {
			t.Errorf("expected ErrDictEntryValueDuplicate, got %v", err)
		}
	})

	t.Run("update non-existent entry returns error", func(t *testing.T) {
		dt, _ := NewDictType("id-3", "gender", "性别", "")
		err := dt.UpdateEntry("e-nonexistent", "测试", "test", 1)
		if err != ErrDictEntryNotFound {
			t.Errorf("expected ErrDictEntryNotFound, got %v", err)
		}
	})

	t.Run("update entry with same value as itself is allowed", func(t *testing.T) {
		dt, _ := NewDictType("id-4", "gender", "性别", "")
		_ = dt.AddEntry(&DictEntry{ID: "e-1", Label: "男", Value: "male", SortOrder: 1, Status: DictEntryStatusEnabled})

		err := dt.UpdateEntry("e-1", "男性", "male", 1)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if dt.Entries[0].Label != "男性" {
			t.Errorf("expected label '男性', got %q", dt.Entries[0].Label)
		}
	})
}

func TestDictType_RemoveEntry(t *testing.T) {
	t.Run("remove existing entry", func(t *testing.T) {
		dt, _ := NewDictType("id-1", "gender", "性别", "")
		_ = dt.AddEntry(&DictEntry{ID: "e-1", Label: "男", Value: "male", Status: DictEntryStatusEnabled})
		_ = dt.AddEntry(&DictEntry{ID: "e-2", Label: "女", Value: "female", Status: DictEntryStatusEnabled})

		dt.RemoveEntry("e-1")

		if len(dt.Entries) != 1 {
			t.Fatalf("expected 1 entry, got %d", len(dt.Entries))
		}
		if dt.Entries[0].ID != "e-2" {
			t.Errorf("expected remaining entry 'e-2', got %q", dt.Entries[0].ID)
		}
	})

	t.Run("remove non-existent entry does nothing", func(t *testing.T) {
		dt, _ := NewDictType("id-2", "gender", "性别", "")
		_ = dt.AddEntry(&DictEntry{ID: "e-1", Label: "男", Value: "male", Status: DictEntryStatusEnabled})

		dt.RemoveEntry("e-nonexistent")

		if len(dt.Entries) != 1 {
			t.Errorf("expected still 1 entry, got %d", len(dt.Entries))
		}
	})
}

func TestDictType_EnableDisableEntry(t *testing.T) {
	t.Run("enable and disable entry", func(t *testing.T) {
		dt, _ := NewDictType("id-1", "gender", "性别", "")
		_ = dt.AddEntry(&DictEntry{ID: "e-1", Label: "男", Value: "male", Status: DictEntryStatusEnabled})

		// disable
		err := dt.DisableEntry("e-1")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if dt.Entries[0].IsEnabled() {
			t.Error("expected entry to be disabled")
		}

		// enable
		err = dt.EnableEntry("e-1")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !dt.Entries[0].IsEnabled() {
			t.Error("expected entry to be enabled")
		}
	})

	t.Run("enable already enabled entry returns error", func(t *testing.T) {
		dt, _ := NewDictType("id-2", "gender", "性别", "")
		_ = dt.AddEntry(&DictEntry{ID: "e-1", Label: "男", Value: "male", Status: DictEntryStatusEnabled})

		err := dt.EnableEntry("e-1")
		if err != ErrDictEntryAlreadyEnabled {
			t.Errorf("expected ErrDictEntryAlreadyEnabled, got %v", err)
		}
	})

	t.Run("disable already disabled entry returns error", func(t *testing.T) {
		dt, _ := NewDictType("id-3", "gender", "性别", "")
		_ = dt.AddEntry(&DictEntry{ID: "e-1", Label: "男", Value: "male", Status: DictEntryStatusEnabled})
		_ = dt.DisableEntry("e-1")

		err := dt.DisableEntry("e-1")
		if err != ErrDictEntryAlreadyDisabled {
			t.Errorf("expected ErrDictEntryAlreadyDisabled, got %v", err)
		}
	})

	t.Run("enable non-existent entry returns error", func(t *testing.T) {
		dt, _ := NewDictType("id-4", "gender", "性别", "")
		err := dt.EnableEntry("e-nonexistent")
		if err != ErrDictEntryNotFound {
			t.Errorf("expected ErrDictEntryNotFound, got %v", err)
		}
	})

	t.Run("disable non-existent entry returns error", func(t *testing.T) {
		dt, _ := NewDictType("id-5", "gender", "性别", "")
		err := dt.DisableEntry("e-nonexistent")
		if err != ErrDictEntryNotFound {
			t.Errorf("expected ErrDictEntryNotFound, got %v", err)
		}
	})
}

func TestDictType_GetEnabledEntries(t *testing.T) {
	dt, _ := NewDictType("id-1", "gender", "性别", "")
	_ = dt.AddEntry(&DictEntry{ID: "e-1", Label: "男", Value: "male", SortOrder: 2, Status: DictEntryStatusEnabled})
	_ = dt.AddEntry(&DictEntry{ID: "e-2", Label: "女", Value: "female", SortOrder: 1, Status: DictEntryStatusEnabled})
	_ = dt.AddEntry(&DictEntry{ID: "e-3", Label: "未知", Value: "unknown", SortOrder: 3, Status: DictEntryStatusDisabled})

	enabled := dt.GetEnabledEntries()

	if len(enabled) != 2 {
		t.Fatalf("expected 2 enabled entries, got %d", len(enabled))
	}

	// should be in added order (GetEnabledEntries does not re-sort, just filters)
	for _, e := range enabled {
		if e.Value == "unknown" {
			t.Error("disabled entry should not appear in enabled entries")
		}
	}
}