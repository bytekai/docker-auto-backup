package test

import (
	"context"
	"reflect"
	"testing"

	"github.com/bytekai/docker-auto-backup/internal/utils"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type mockClient struct {
	client.Client
	inspectFunc func(ctx context.Context, containerID string) (types.ContainerJSON, error)
}

func (m *mockClient) ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error) {
	return m.inspectFunc(ctx, containerID)
}

func TestGetContainerEnv(t *testing.T) {
	tests := []struct {
		name        string
		mockInspect func(ctx context.Context, containerID string) (types.ContainerJSON, error)
		want        map[string]string
		wantErr     bool
	}{
		{
			name: "successful env retrieval",
			mockInspect: func(ctx context.Context, containerID string) (types.ContainerJSON, error) {
				return types.ContainerJSON{
					ContainerJSONBase: &types.ContainerJSONBase{},
					Config: &container.Config{
						Env: []string{"KEY1=value1", "KEY2=value2"},
					},
				}, nil
			},
			want: map[string]string{
				"KEY1": "value1",
				"KEY2": "value2",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cli := &mockClient{inspectFunc: tt.mockInspect}
			got, err := utils.GetContainerEnv(context.Background(), cli, &types.Container{ID: "test-container"})

			if (err != nil) != tt.wantErr {
				t.Errorf("getContainerEnv() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getContainerEnv() = %v, want %v", got, tt.want)
			}
		})
	}
}
