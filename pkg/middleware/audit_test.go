package middleware

import (
	"testing"
)

func TestGetActionFromMethod(t *testing.T) {
	tests := []struct {
		method   string
		expected string
	}{
		{"POST", "create"},
		{"PUT", "update"},
		{"DELETE", "delete"},
		{"GET", ""},
		{"PATCH", ""},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			result := getActionFromMethod(tt.method)
			if result != tt.expected {
				t.Errorf("getActionFromMethod(%q) = %q, want %q", tt.method, result, tt.expected)
			}
		})
	}
}

func TestInferTargetType(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"/api/v1/roles", "role"},
		{"/api/v1/roles/123", "role"},
		{"/api/v1/permissions", "permission"},
		{"/api/v1/users", "user"},
		{"/api/v1/dict-types", "dict_type"},
		{"/api/v1/dict-entries", "dict_entry"},
		{"/api/v1/kv-configs", "kv_config"},
		{"/api/v1/unknown", "unknown"},
		{"/api/v1/", ""},
		{"/other/path", ""},
		{"/api/v1/roles/123/permissions", "role"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := inferTargetType(tt.path)
			if result != tt.expected {
				t.Errorf("inferTargetType(%q) = %q, want %q", tt.path, result, tt.expected)
			}
		})
	}
}

func TestExtractIDFromResponse(t *testing.T) {
	t.Run("standard response", func(t *testing.T) {
		body := `{"code":0,"message":"ok","data":{"id":"role-123"}}`
		id := extractIDFromResponse(body)
		if id != "role-123" {
			t.Errorf("expected 'role-123', got %q", id)
		}
	})

	t.Run("nested data", func(t *testing.T) {
		body := `{"code":0,"message":"ok","data":{"id":"user-456","name":"test"}}`
		id := extractIDFromResponse(body)
		if id != "user-456" {
			t.Errorf("expected 'user-456', got %q", id)
		}
	})

	t.Run("empty body", func(t *testing.T) {
		id := extractIDFromResponse("")
		if id != "" {
			t.Errorf("expected empty, got %q", id)
		}
	})

	t.Run("no id field", func(t *testing.T) {
		body := `{"code":0,"message":"ok","data":{"name":"test"}}`
		id := extractIDFromResponse(body)
		if id != "" {
			t.Errorf("expected empty, got %q", id)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		body := `not json`
		id := extractIDFromResponse(body)
		if id != "" {
			t.Errorf("expected empty, got %q", id)
		}
	})
}

func TestTruncateString(t *testing.T) {
	t.Run("shorter than max", func(t *testing.T) {
		result := truncateString("hello", 10)
		if result != "hello" {
			t.Errorf("expected 'hello', got %q", result)
		}
	})

	t.Run("equal to max", func(t *testing.T) {
		result := truncateString("hello", 5)
		if result != "hello" {
			t.Errorf("expected 'hello', got %q", result)
		}
	})

	t.Run("longer than max", func(t *testing.T) {
		result := truncateString("hello world", 5)
		if result != "hello" {
			t.Errorf("expected 'hello', got %q", result)
		}
	})

	t.Run("empty string", func(t *testing.T) {
		result := truncateString("", 10)
		if result != "" {
			t.Errorf("expected empty, got %q", result)
		}
	})
}