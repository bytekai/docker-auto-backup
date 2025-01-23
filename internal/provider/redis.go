package provider

import (
	"context"

	"github.com/bytekai/docker-auto-backup/internal/logger"
	"github.com/bytekai/docker-auto-backup/internal/models"
	"github.com/docker/docker/client"
)

type RedisProviderConfig struct{}

type RedisProvider struct {
	name   string
	images []string
	ext    string
	ctx    *ProviderContext
}

func NewRedisProvider(ctx *ProviderContext) *RedisProvider {
	return &RedisProvider{
		name:   "redis",
		images: []string{"redis"},
		ext:    "dump",
		ctx:    ctx,
	}
}

func (p *RedisProvider) Backup(c context.Context, storage models.Storage) error {
	return nil
}

func (p *RedisProvider) Restore(c context.Context, session *logger.Session, cli *client.Client, backupPath string, containerID string) error {
	return nil
}
