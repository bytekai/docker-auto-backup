package test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bytekai/docker-auto-backup/internal/storage"
)

func TestLocalStorage_Put(t *testing.T) {
	tempDir := t.TempDir()
	storage := storage.NewLocalStorage(storage.LocalStorageConfig{RootPath: tempDir})

	t.Run("successful file upload", func(t *testing.T) {
		content := "test content"
		reader := strings.NewReader(content)

		err := storage.Put(context.Background(), "test.tar.gz", reader)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := storage.Put(ctx, "test.tar.gz", strings.NewReader("test"))
		if err != context.Canceled {
			t.Errorf("expected context.Canceled error, got %v", err)
		}
	})
}

func TestLocalStorage_Get(t *testing.T) {
	tempDir := t.TempDir()
	storage := storage.NewLocalStorage(storage.LocalStorageConfig{RootPath: tempDir})

	t.Run("successful file retrieval", func(t *testing.T) {
		content := "test content"
		filename := "backup-test.tar.gz"
		fullPath := filepath.Join(tempDir, filename)

		err := os.WriteFile(fullPath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		reader, err := storage.Get(context.Background(), filename)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer reader.Close()

		buf := new(bytes.Buffer)
		_, err = io.Copy(buf, reader)
		if err != nil {
			t.Fatalf("failed to read file content: %v", err)
		}

		if buf.String() != content {
			t.Errorf("content mismatch: expected %q, got %q", content, buf.String())
		}
	})

	t.Run("file not found", func(t *testing.T) {
		_, err := storage.Get(context.Background(), "nonexistent.tar.gz")
		if !errors.Is(err, os.ErrNotExist) {
			t.Errorf("expected os.IsNotExist error, got %v", err)
		}
	})

	t.Run("cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := storage.Get(ctx, "test.tar.gz")
		if err != context.Canceled {
			t.Errorf("expected context.Canceled error, got %v", err)
		}
	})
}
