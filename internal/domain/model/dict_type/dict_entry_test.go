package dict_type

import (
	"testing"
)

func TestDictEntry_Enable(t *testing.T) {
	e := &DictEntry{Status: DictEntryStatusDisabled}

	e.Enable()

	if e.Status != DictEntryStatusEnabled {
		t.Errorf("expected status %d (enabled), got %d", DictEntryStatusEnabled, e.Status)
	}
	if e.UpdatedAt.IsZero() {
		t.Error("expected non-zero UpdatedAt after enable")
	}
}

func TestDictEntry_Disable(t *testing.T) {
	e := &DictEntry{Status: DictEntryStatusEnabled}

	e.Disable()

	if e.Status != DictEntryStatusDisabled {
		t.Errorf("expected status %d (disabled), got %d", DictEntryStatusDisabled, e.Status)
	}
	if e.UpdatedAt.IsZero() {
		t.Error("expected non-zero UpdatedAt after disable")
	}
}

func TestDictEntry_ChangeSort(t *testing.T) {
	e := &DictEntry{SortOrder: 1}

	e.ChangeSort(5)

	if e.SortOrder != 5 {
		t.Errorf("expected SortOrder 5, got %d", e.SortOrder)
	}
	if e.UpdatedAt.IsZero() {
		t.Error("expected non-zero UpdatedAt after change sort")
	}
}

func TestDictEntry_IsEnabled(t *testing.T) {
	t.Run("enabled returns true", func(t *testing.T) {
		e := &DictEntry{Status: DictEntryStatusEnabled}
		if !e.IsEnabled() {
			t.Error("expected IsEnabled to be true")
		}
	})

	t.Run("disabled returns false", func(t *testing.T) {
		e := &DictEntry{Status: DictEntryStatusDisabled}
		if e.IsEnabled() {
			t.Error("expected IsEnabled to be false")
		}
	})
}