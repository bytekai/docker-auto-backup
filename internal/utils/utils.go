package utils

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func GetContainerEnv(ctx context.Context, cli client.APIClient, container *types.Container) (map[string]string, error) {
	env := make(map[string]string)
	cInfo, err := cli.ContainerInspect(ctx, container.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect container: %w", err)
	}

	for _, e := range cInfo.Config.Env {
		key, value := parseEnvVar(e)
		if key != "" {
			env[key] = value
		}
	}

	return env, nil
}

func parseEnvVar(env string) (key, value string) {
	parts := strings.SplitN(env, "=", 2)

	if len(parts) != 2 {
		return "", ""
	}

	key = strings.TrimSpace(parts[0])
	value = strings.TrimSpace(parts[1])

	return key, value
}
