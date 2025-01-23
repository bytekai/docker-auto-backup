package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bytekai/docker-auto-backup/internal/logger"
	"github.com/bytekai/docker-auto-backup/internal/manager"
	"github.com/bytekai/docker-auto-backup/internal/provider"
	"github.com/bytekai/docker-auto-backup/internal/scheduler"
	"github.com/bytekai/docker-auto-backup/internal/storage"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

func extractLabels(labels map[string]string) map[string]string {
	extracted := make(map[string]string)
	for key, value := range labels {
		if strings.HasPrefix(key, "backup.") {
			extracted[strings.TrimPrefix(key, "backup.")] = value
		}
	}
	return extracted
}

func validateFrequency(freq string) (scheduler.Frequency, error) {
	validFrequencies := map[string]scheduler.Frequency{
		"daily":   scheduler.Daily,
		"weekly":  scheduler.Weekly,
		"monthly": scheduler.Monthly,
		"yearly":  scheduler.Yearly,
	}

	if frequency, ok := validFrequencies[freq]; ok {
		return frequency, nil
	}
	return "", fmt.Errorf("invalid frequency: %s", freq)
}

func getStringWithDefault(labels map[string]string, key, defaultValue string) string {
	if val := labels[key]; val != "" {
		return val
	}
	return defaultValue
}

func parseIntWithDefault(labels map[string]string, key string, defaultValue int) (int, error) {
	if val := labels[key]; val != "" {
		return strconv.Atoi(val)
	}
	return defaultValue, nil
}

func buildStorageConfig(labels map[string]string) (*storage.StorageConfig, error) {
	storageType := getStringWithDefault(labels, "storage", "local")

	storageConfig := &storage.StorageConfig{}
	switch storageType {
	case "local":
		storageConfig.Local = &storage.LocalStorageConfig{
			RootPath: labels["storage.local.root_path"],
		}
	case "s3":
		storageConfig.S3 = &storage.S3StorageConfig{
			Bucket:    labels["storage.s3.bucket"],
			Region:    labels["storage.s3.region"],
			AccessKey: labels["storage.s3.access_key"],
			SecretKey: labels["storage.s3.secret_key"],
		}
	default:
		return nil, fmt.Errorf("invalid storage: %s", storageType)
	}

	return storageConfig, nil
}

func buildProviderConfig(labels map[string]string) *provider.ProviderConfig {
	providerType := getStringWithDefault(labels, "provider", "local")

	providerConfig := &provider.ProviderConfig{}
	switch providerType {
	case "postgres":
		providerConfig.Postgres = &provider.PostgresProviderConfig{}
	case "redis":
		providerConfig.Redis = &provider.RedisProviderConfig{}
	case "clickhouse":
		providerConfig.Clickhouse = &provider.ClickhouseProviderConfig{}
	case "nats":
		providerConfig.Nats = &provider.NatsProviderConfig{}
	case "rabbitmq":
		providerConfig.RabbitMQ = &provider.RabbitMQProviderConfig{}
	}

	return providerConfig
}

func parseConfig(labels map[string]string) (*scheduler.Config, error) {
	if labels["enabled"] != "true" {
		return &scheduler.Config{Enabled: false}, nil
	}

	timeZone := getStringWithDefault(labels, "time_zone", "UTC")
	tz, err := time.LoadLocation(timeZone)
	if err != nil {
		return nil, fmt.Errorf("failed to load time zone: %v", err)
	}

	frequency, err := validateFrequency(labels["frequency"])
	if err != nil {
		return nil, err
	}

	dayOfMonth, err := parseIntWithDefault(labels, "day_of_month", 1)
	if err != nil {
		return nil, fmt.Errorf("failed to parse day of month: %v", err)
	}

	dayOfWeek, err := parseIntWithDefault(labels, "day_of_week", 0)
	if err != nil {
		return nil, fmt.Errorf("failed to parse day of week: %v", err)
	}

	dayOfYear, err := parseIntWithDefault(labels, "day_of_year", 1)
	if err != nil {
		return nil, fmt.Errorf("failed to parse day of year: %v", err)
	}

	storageConfig, err := buildStorageConfig(labels)
	if err != nil {
		return nil, err
	}

	return &scheduler.Config{
		Enabled:        true,
		Frequency:      frequency,
		Time:           getStringWithDefault(labels, "time", "00:00"),
		TimeZone:       tz,
		DayOfMonth:     dayOfMonth,
		DayOfWeek:      dayOfWeek,
		DayOfYear:      dayOfYear,
		Provider:       getStringWithDefault(labels, "provider", "local"),
		Location:       getStringWithDefault(labels, "location", "local"),
		StorageConfig:  storageConfig,
		ProviderConfig: buildProviderConfig(labels),
	}, nil
}

func main() {
	log := logger.New(logger.DEBUG)
	session := log.NewSession("[main] ")

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		session.Error("Failed to create Docker client: %v", err)
		os.Exit(1)
	}

	mgr := manager.New(log)

	ctx := context.Background()
	containers, err := cli.ContainerList(ctx, container.ListOptions{})
	if err != nil {
		session.Error("Failed to list containers: %v", err)
		os.Exit(1)
	}

	for _, container := range containers {
		labels := extractLabels(container.Labels)
		if err := handleContainer(ctx, cli, container.ID, labels, mgr, log); err != nil {
			session.Error("Failed to handle container %s: %v", container.ID, err)
		}
	}

	filterArgs := filters.NewArgs()
	filterArgs.Add("type", "container")
	filterArgs.Add("event", "start")
	filterArgs.Add("event", "die")

	eventsCh, errCh := cli.Events(ctx, events.ListOptions{
		Filters: filterArgs,
	})

	for {
		select {
		case event := <-eventsCh:
			containerID := event.Actor.ID
			labels := extractLabels(event.Actor.Attributes)

			switch event.Action {
			case "start":
				if err := handleContainer(ctx, cli, containerID, labels, mgr, log); err != nil {
					session.Error("Failed to handle container start %s: %v", containerID, err)
				}
			case "die":
				mgr.RemoveScheduler(containerID)
				session.Info("Removed scheduler for container: %s", containerID)
			}

		case err := <-errCh:
			session.Error("Error watching events: %v", err)
			time.Sleep(5 * time.Second)
		}
	}
}

func handleContainer(ctx context.Context, cli *client.Client, containerID string, labels map[string]string, mgr *manager.Manager, log *logger.Logger) error {
	config, err := parseConfig(labels)
	if err != nil {
		return fmt.Errorf("failed to parse config: %v", err)
	}

	if !config.Enabled {
		return nil
	}

	session := log.NewSession("[backup] ")

	json, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	session.Info("Config for container %s: %s", containerID, string(json))

	sch, err := scheduler.New(*config, func() {
		session.Info("Executing backup task for container: %s", containerID)

		pCtx := &provider.ProviderContext{
			Session:     session,
			Client:      cli,
			ContainerID: containerID,
		}

		p := provider.NewProvider(pCtx, config.Provider)

		storage := storage.NewStorage(pCtx, config.Location, config.StorageConfig)

		if err := p.Backup(ctx, storage); err != nil {
			session.Error("Failed to backup container %s: %v", containerID, err)
		}
	}, scheduler.WithLogger(log))
	if err != nil {
		return fmt.Errorf("failed to create scheduler: %v", err)
	}

	if err := sch.Start(); err != nil {
		return fmt.Errorf("failed to start scheduler: %v", err)
	}

	mgr.AddScheduler(containerID, sch)
	return nil
}
