package provider

import (
	"context"

	"github.com/bytekai/docker-auto-backup/internal/logger"
	"github.com/bytekai/docker-auto-backup/internal/models"
	"github.com/docker/docker/client"
)

type Provider interface {
	Backup(ctx context.Context, storage models.Storage) error
	Restore(ctx context.Context, session *logger.Session, cli *client.Client, backupPath string, containerID string) error
}

type ProviderContext struct {
	Session     *logger.Session
	Client      *client.Client
	ContainerID string
}

type ProviderConfig struct {
	Postgres   *PostgresProviderConfig
	Redis      *RedisProviderConfig
	Clickhouse *ClickhouseProviderConfig
	Nats       *NatsProviderConfig
	RabbitMQ   *RabbitMQProviderConfig
}

func NewProvider(ctx *ProviderContext, provider string) Provider {
	switch provider {
	case "postgres":
		return &PostgresProvider{ctx: ctx}
	case "redis":
		return &RedisProvider{ctx: ctx}
	case "clickhouse":
		return &ClickhouseProvider{ctx: ctx}
	case "nats":
		return &NatsProvider{ctx: ctx}
	case "rabbitmq":
		return &RabbitMQProvider{ctx: ctx}
	default:
		return nil
	}
}
