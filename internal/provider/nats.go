package provider

import (
	"context"

	"github.com/bytekai/docker-auto-backup/internal/logger"
	"github.com/bytekai/docker-auto-backup/internal/models"
	"github.com/docker/docker/client"
)

type NatsProviderConfig struct{}
type NatsProvider struct {
	ctx *ProviderContext
}

func NewNatsProvider(ctx *ProviderContext) *NatsProvider {
	return &NatsProvider{ctx: ctx}
}

func (p *NatsProvider) Backup(c context.Context, storage models.Storage) error {
	return nil
}

func (p *NatsProvider) Restore(c context.Context, session *logger.Session, cli *client.Client, backupPath string, containerID string) error {
	return nil
}
