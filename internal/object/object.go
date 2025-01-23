package backup

import (
	"context"
	"fmt"
	"io"
	"time"
)

type BackupObject interface {
	GetExt() string
	GetTimestamp() string
	GetData() io.Reader
}

type backupObject struct {
	extension string
	createdAt time.Time
	size      int64
	metadata  map[string]string
	reader    io.ReadCloser
}

type Option func(*backupObject)

func WithMetadata(metadata map[string]string) Option {
	return func(o *backupObject) {
		o.metadata = metadata
	}
}

func WithSize(size int64) Option {
	return func(o *backupObject) {
		o.size = size
	}
}

func New(ctx context.Context, ext string, reader io.ReadCloser, opts ...Option) (*backupObject, error) {
	if ctx.Err() != nil {
		return nil, fmt.Errorf("context error: %w", ctx.Err())
	}

	if ext == "" {
		return nil, fmt.Errorf("extension cannot be empty")
	}

	if reader == nil {
		return nil, fmt.Errorf("reader cannot be nil")
	}

	obj := &backupObject{
		extension: ext,
		createdAt: time.Now(),
		reader:    reader,
		metadata:  make(map[string]string),
	}

	for _, opt := range opts {
		opt(obj)
	}

	return obj, nil
}
