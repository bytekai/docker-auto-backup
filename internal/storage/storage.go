package storage

import (
	"github.com/bytekai/docker-auto-backup/internal/models"
	"github.com/bytekai/docker-auto-backup/internal/provider"
)

type StorageConfig struct {
	Local *LocalStorageConfig
	S3    *S3StorageConfig
}

func NewStorage(ctx *provider.ProviderContext, provider string, config *StorageConfig) models.Storage {
	switch provider {
	case "local":
		if config.Local == nil {
			ctx.Session.Error("Local storage configuration is missing")
			return nil
		}
		return &LocalStorage{config: *config.Local}
	case "s3":
		if config.S3 == nil {
			ctx.Session.Error("S3 storage configuration is missing")
			return nil
		}

		return &S3Storage{config: *config.S3}
	default:
		ctx.Session.Error("Unsupported provider: %s", provider)
		return nil
	}
}
