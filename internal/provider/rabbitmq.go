package provider

import (
	"context"

	"github.com/bytekai/docker-auto-backup/internal/logger"
	"github.com/bytekai/docker-auto-backup/internal/models"
	"github.com/docker/docker/client"
)

type RabbitMQProvider struct {
	ctx *ProviderContext
}

type RabbitMQProviderConfig struct{}

func NewRabbitMQProvider(ctx *ProviderContext) *RabbitMQProvider {
	return &RabbitMQProvider{ctx: ctx}
}

func (p *RabbitMQProvider) Backup(c context.Context, storage models.Storage) error {
	return nil
}

func (p *RabbitMQProvider) Restore(c context.Context, session *logger.Session, cli *client.Client, backupPath string, containerID string) error {
	return nil
}
