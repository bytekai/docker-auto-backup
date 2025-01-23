package models

import (
	"context"
	"io"
)

type Storage interface {
	Put(ctx context.Context, name string, file io.Reader) error
	Get(ctx context.Context, name string) (io.ReadCloser, error)
}
