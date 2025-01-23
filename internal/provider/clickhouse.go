package provider

import (
	"context"

	"github.com/bytekai/docker-auto-backup/internal/logger"
	"github.com/bytekai/docker-auto-backup/internal/models"
	"github.com/docker/docker/client"
)

type ClickhouseProviderConfig struct{}

type ClickhouseProvider struct {
	ctx *ProviderContext
}

func NewClickhouseProvider(ctx *ProviderContext) *ClickhouseProvider {
	return &ClickhouseProvider{ctx: ctx}
}

func (p *ClickhouseProvider) Backup(c context.Context, storage models.Storage) error {
	return nil
}

func (p *ClickhouseProvider) Restore(c context.Context, session *logger.Session, cli *client.Client, backupPath string, containerID string) error {
	return nil
}
