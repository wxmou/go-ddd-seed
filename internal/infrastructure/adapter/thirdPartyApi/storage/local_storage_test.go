package storage

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLocalStorage_Upload(t *testing.T) {
	t.Parallel()

	type args struct {
		content string
		path    string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "upload file to root",
			args: args{
				content: "hello world",
				path:    "test.txt",
			},
			wantErr: false,
		},
		{
			name: "upload file to nested directory",
			args: args{
				content: "nested content",
				path:    "sub/dir/file.txt",
			},
			wantErr: false,
		},
		{
			name: "upload empty file",
			args: args{
				content: "",
				path:    "empty.txt",
			},
			wantErr: false,
		},
		{
			name: "upload file with Chinese filename",
			args: args{
				content: "中文内容",
				path:    "中文/文件.txt",
			},
			wantErr: false,
		},
		{
			name: "upload file with special characters in path",
			args: args{
				content: "special",
				path:    "path/with spaces/file-1.2_test.txt",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			basePath := t.TempDir()
			store := NewLocalStorage(basePath, "http://localhost:8080")

			reader := strings.NewReader(tt.args.content)
			err := store.Upload(context.Background(), reader, tt.args.path)

			if (err != nil) != tt.wantErr {
				t.Fatalf("Upload() error = %v, wantErr = %v", err, tt.wantErr)
			}

			// Verify file was written correctly
			fullPath := filepath.Join(basePath, tt.args.path)
			data, err := os.ReadFile(fullPath)
			if err != nil {
				t.Fatalf("failed to read uploaded file: %v", err)
			}
			if string(data) != tt.args.content {
				t.Errorf("file content = %q, want %q", string(data), tt.args.content)
			}
		})
	}
}

func TestLocalStorage_Download(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setupFunc  func(basePath string) // prepare the file before download
		path       string
		wantContent string
		wantErr    bool
	}{
		{
			name: "download existing file",
			setupFunc: func(basePath string) {
				store := NewLocalStorage(basePath, "")
				store.Upload(context.Background(), strings.NewReader("hello"), "file.txt")
			},
			path:         "file.txt",
			wantContent:  "hello",
			wantErr:      false,
		},
		{
			name: "download file from nested directory",
			setupFunc: func(basePath string) {
				store := NewLocalStorage(basePath, "")
				store.Upload(context.Background(), strings.NewReader("nested"), "a/b/c/nested.txt")
			},
			path:         "a/b/c/nested.txt",
			wantContent:  "nested",
			wantErr:      false,
		},
		{
			name: "download non-existent file",
			setupFunc: func(basePath string) {
				// no file created
			},
			path:       "nonexistent.txt",
			wantContent: "",
			wantErr:    true,
		},
		{
			name: "download from non-existent directory",
			setupFunc: func(basePath string) {
				// no file created
			},
			path:       "missing/dir/file.txt",
			wantContent: "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			basePath := t.TempDir()
			tt.setupFunc(basePath)

			store := NewLocalStorage(basePath, "")
			rc, err := store.Download(context.Background(), tt.path)

			if (err != nil) != tt.wantErr {
				t.Fatalf("Download() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			defer rc.Close()

			data, err := io.ReadAll(rc)
			if err != nil {
				t.Fatalf("failed to read downloaded content: %v", err)
			}
			if string(data) != tt.wantContent {
				t.Errorf("downloaded content = %q, want %q", string(data), tt.wantContent)
			}
		})
	}
}

func TestLocalStorage_Delete(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setupFunc func(basePath string)
		path      string
		wantErr   bool
	}{
		{
			name: "delete existing file",
			setupFunc: func(basePath string) {
				store := NewLocalStorage(basePath, "")
				store.Upload(context.Background(), strings.NewReader("content"), "delete-me.txt")
			},
			path:    "delete-me.txt",
			wantErr: false,
		},
		{
			name: "delete file in nested directory",
			setupFunc: func(basePath string) {
				store := NewLocalStorage(basePath, "")
				store.Upload(context.Background(), strings.NewReader("content"), "deep/nested/file.txt")
			},
			path:    "deep/nested/file.txt",
			wantErr: false,
		},
		{
			name: "delete non-existent file",
			setupFunc: func(basePath string) {
				// no file created
			},
			path:    "nonexistent.txt",
			wantErr: true,
		},
		{
			name: "delete already deleted file",
			setupFunc: func(basePath string) {
				store := NewLocalStorage(basePath, "")
				store.Upload(context.Background(), strings.NewReader("content"), "temp.txt")
				store.Delete(context.Background(), "temp.txt")
			},
			path:    "temp.txt",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			basePath := t.TempDir()
			tt.setupFunc(basePath)

			store := NewLocalStorage(basePath, "")
			err := store.Delete(context.Background(), tt.path)

			if (err != nil) != tt.wantErr {
				t.Fatalf("Delete() error = %v, wantErr = %v", err, tt.wantErr)
			}

			// Verify file is actually removed from disk
			if err == nil {
				fullPath := filepath.Join(basePath, tt.path)
				if _, statErr := os.ReadFile(fullPath); statErr == nil {
					t.Error("file still exists after Delete()")
				}
			}
		})
	}
}

func TestLocalStorage_GetURL(t *testing.T) {
	t.Parallel()

	type args struct {
		baseURL string
		path    string
		expiry  time.Duration
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "get URL with base URL",
			args: args{
				baseURL: "http://localhost:8080/files",
				path:    "test.txt",
				expiry:  0,
			},
			want:    "http://localhost:8080/files/test.txt",
			wantErr: false,
		},
		{
			name: "get URL with trailing slash in base URL",
			args: args{
				baseURL: "http://localhost:8080/files/",
				path:    "test.txt",
				expiry:  0,
			},
			want:    "http://localhost:8080/files//test.txt",
			wantErr: false,
		},
		{
			name: "get URL with empty base URL",
			args: args{
				baseURL: "",
				path:    "test.txt",
				expiry:  0,
			},
			want:    "/test.txt",
			wantErr: false,
		},
		{
			name: "get URL with nested path",
			args: args{
				baseURL: "https://storage.example.com",
				path:    "folder/sub/file.txt",
				expiry:  0,
			},
			want:    "https://storage.example.com/folder/sub/file.txt",
			wantErr: false,
		},
		{
			name: "get URL ignores expiry (local storage doesn't support signed URLs)",
			args: args{
				baseURL: "http://localhost:8080",
				path:    "file.pdf",
				expiry:  time.Hour,
			},
			want:    "http://localhost:8080/file.pdf",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			basePath := t.TempDir()
			store := NewLocalStorage(basePath, tt.args.baseURL)

			got, err := store.GetURL(context.Background(), tt.args.path, tt.args.expiry)

			if (err != nil) != tt.wantErr {
				t.Fatalf("GetURL() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("GetURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLocalStorage_Exists(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setupFunc func(basePath string)
		path      string
		want      bool
		wantErr   bool
	}{
		{
			name: "existing file",
			setupFunc: func(basePath string) {
				store := NewLocalStorage(basePath, "")
				store.Upload(context.Background(), strings.NewReader("exists"), "exists.txt")
			},
			path:    "exists.txt",
			want:    true,
			wantErr: false,
		},
		{
			name: "non-existent file",
			setupFunc: func(basePath string) {
				// no file created
			},
			path:    "nonexistent.txt",
			want:    false,
			wantErr: false,
		},
		{
			name: "non-existent file in nested path",
			setupFunc: func(basePath string) {
				// no file created
			},
			path:    "deep/nested/missing.txt",
			want:    false,
			wantErr: false,
		},
		{
			name: "file after deletion returns false",
			setupFunc: func(basePath string) {
				store := NewLocalStorage(basePath, "")
				store.Upload(context.Background(), strings.NewReader("temp"), "temp.txt")
				store.Delete(context.Background(), "temp.txt")
			},
			path:    "temp.txt",
			want:    false,
			wantErr: false,
		},
		{
			name: "file after upload and recheck",
			setupFunc: func(basePath string) {
				store := NewLocalStorage(basePath, "")
				store.Upload(context.Background(), strings.NewReader("persistent"), "persistent.txt")
			},
			path:    "persistent.txt",
			want:    true,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			basePath := t.TempDir()
			tt.setupFunc(basePath)

			store := NewLocalStorage(basePath, "")
			got, err := store.Exists(context.Background(), tt.path)

			if (err != nil) != tt.wantErr {
				t.Fatalf("Exists() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("Exists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLocalStorage_UploadAndDownloadRoundtrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
		path    string
	}{
		{
			name:    "roundtrip text file",
			content: "The quick brown fox jumps over the lazy dog.",
			path:    "roundtrip.txt",
		},
		{
			name:    "roundtrip JSON content",
			content: `{"name":"test","value":123,"nested":{"key":true}}`,
			path:    "data/config.json",
		},
		{
			name:    "roundtrip binary-like content",
			content: "line1\nline2\nline3\n",
			path:    "multiline/data.txt",
		},
		{
			name:    "roundtrip empty content",
			content: "",
			path:    "empty.txt",
		},
		{
			name:    "roundtrip large content",
			content: strings.Repeat("A", 10000),
			path:    "large/file.bin",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			basePath := t.TempDir()
			store := NewLocalStorage(basePath, "")

			// Upload
			err := store.Upload(context.Background(), strings.NewReader(tt.content), tt.path)
			if err != nil {
				t.Fatalf("Upload() failed: %v", err)
			}

			// Verify Exists returns true
			exists, err := store.Exists(context.Background(), tt.path)
			if err != nil {
				t.Fatalf("Exists() after upload failed: %v", err)
			}
			if !exists {
				t.Error("Exists() returned false after successful upload")
			}

			// Download and verify content
			rc, err := store.Download(context.Background(), tt.path)
			if err != nil {
				t.Fatalf("Download() failed: %v", err)
			}

			data, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				t.Fatalf("failed to read downloaded content: %v", err)
			}
			if string(data) != tt.content {
				t.Errorf("downloaded content = %q (len=%d), want %q (len=%d)",
					string(data), len(data), tt.content, len(tt.content))
			}

			// GetURL should return a non-empty URL
			url, err := store.GetURL(context.Background(), tt.path, 0)
			if err != nil {
				t.Fatalf("GetURL() failed: %v", err)
			}
			if url == "" {
				t.Error("GetURL() returned empty URL")
			}

			// Delete the file
			err = store.Delete(context.Background(), tt.path)
			if err != nil {
				t.Fatalf("Delete() failed: %v", err)
			}

			// Verify Exists returns false after delete
			exists, err = store.Exists(context.Background(), tt.path)
			if err != nil {
				t.Fatalf("Exists() after delete failed: %v", err)
			}
			if exists {
				t.Error("Exists() returned true after successful delete")
			}
		})
	}
}

func TestLocalStorage_Constructor(t *testing.T) {
	t.Parallel()

	t.Run("creates base directory", func(t *testing.T) {
		basePath := filepath.Join(t.TempDir(), "auto-created")
		store := NewLocalStorage(basePath, "")

		// The constructor should create the base directory
		if store == nil {
			t.Fatal("NewLocalStorage returned nil")
		}
		if store.basePath != basePath {
			t.Errorf("basePath = %q, want %q", store.basePath, basePath)
		}
	})

	t.Run("reuses existing directory", func(t *testing.T) {
		basePath := t.TempDir()
		store := NewLocalStorage(basePath, "http://example.com")

		if store == nil {
			t.Fatal("NewLocalStorage returned nil")
		}
		if store.baseURL != "http://example.com" {
			t.Errorf("baseURL = %q, want %q", store.baseURL, "http://example.com")
		}
	})
}

func TestLocalStorage_UploadWithReaderType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content []byte
		path    string
	}{
		{
			name:    "upload with bytes.Reader",
			content: []byte("bytes reader content"),
			path:    "bytes-reader.txt",
		},
		{
			name:    "upload with buffer",
			content: []byte{0x00, 0x01, 0x02, 0xFF},
			path:    "binary.bin",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			basePath := t.TempDir()
			store := NewLocalStorage(basePath, "")

			err := store.Upload(context.Background(), bytes.NewReader(tt.content), tt.path)
			if err != nil {
				t.Fatalf("Upload() failed: %v", err)
			}

			// Download and verify
			rc, err := store.Download(context.Background(), tt.path)
			if err != nil {
				t.Fatalf("Download() failed: %v", err)
			}
			defer rc.Close()

			got, err := io.ReadAll(rc)
			if err != nil {
				t.Fatalf("failed to read: %v", err)
			}
			if !bytes.Equal(got, tt.content) {
				t.Errorf("content mismatch: got %v, want %v", got, tt.content)
			}
		})
	}
}