package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type LocalStorage struct {
	config LocalStorageConfig
}

type LocalStorageConfig struct {
	RootPath string
}

func NewLocalStorage(config LocalStorageConfig) LocalStorage {
	return LocalStorage{config: config}
}

func (s *LocalStorage) Put(ctx context.Context, name string, file io.Reader) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	fullPath := filepath.Join(s.config.RootPath, name)

	err := os.MkdirAll(filepath.Dir(fullPath), os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	out, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func (s *LocalStorage) Get(ctx context.Context, url string) (io.ReadCloser, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	fullPath := filepath.Join(s.config.RootPath, url)
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return file, nil
}
