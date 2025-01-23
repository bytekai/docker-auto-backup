package provider

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/bytekai/docker-auto-backup/internal/logger"
	"github.com/bytekai/docker-auto-backup/internal/models"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type PostgresProviderConfig struct{}

type PostgresProvider struct {
	name   string
	images []string
	ext    string
	ctx    *ProviderContext
}

func NewPostgresProvider(ctx *ProviderContext) *PostgresProvider {
	return &PostgresProvider{
		name:   "postgres",
		images: []string{"postgres"},
		ext:    "sql",
		ctx:    ctx,
	}
}

func (p *PostgresProvider) Backup(c context.Context, storage models.Storage) error {
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("backup_%s.%s", timestamp, p.ext)
	outputPath := filepath.Join("/backups", filename)

	p.ctx.Session.Info("Backing up container %s to %s", p.ctx.ContainerID, outputPath)

	execConfig := container.ExecOptions{
		Cmd: []string{
			"pg_dumpall",
			"-U", "postgres",
			"--clean",
		},
		AttachStdout: true,
		AttachStderr: true,
	}

	execResp, err := p.ctx.Client.ContainerExecCreate(c, p.ctx.ContainerID, execConfig)
	if err != nil {
		p.ctx.Session.Error("Failed to create exec: %v", err)
		return fmt.Errorf("failed to create exec: %w", err)
	}

	resp, err := p.ctx.Client.ContainerExecAttach(c, execResp.ID, container.ExecStartOptions{})
	if err != nil {
		p.ctx.Session.Error("Failed to attach to exec: %v", err)
		return fmt.Errorf("failed to attach to exec: %w", err)
	}
	defer resp.Close()

	err = storage.Put(c, filename, resp.Reader)
	if err != nil {
		p.ctx.Session.Error("Failed to store backup: %v", err)
		return fmt.Errorf("failed to store backup: %w", err)
	}

	execInspect, err := p.ctx.Client.ContainerExecInspect(c, execResp.ID)
	if err != nil {
		p.ctx.Session.Error("Failed to inspect exec: %v", err)
		return fmt.Errorf("failed to inspect exec: %w", err)
	}

	if execInspect.ExitCode != 0 {
		p.ctx.Session.Error("pg_dump failed with exit code %d", execInspect.ExitCode)
		return fmt.Errorf("pg_dump failed with exit code %d", execInspect.ExitCode)
	}

	return nil
}

func (p *PostgresProvider) Restore(ctx context.Context, session *logger.Session, cli *client.Client, backupPath string, containerID string) error {
	file, err := os.Open(backupPath)
	if err != nil {
		session.Error("Failed to open backup file: %v", err)
		return fmt.Errorf("failed to open backup file: %w", err)
	}
	defer file.Close()

	execConfig := container.ExecOptions{
		Cmd: []string{
			"psql",
			"-U", "postgres",
			"-f", "-",
		},
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
	}

	execResp, err := cli.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		session.Error("Failed to create exec: %v", err)
		return fmt.Errorf("failed to create exec: %w", err)
	}

	resp, err := cli.ContainerExecAttach(ctx, execResp.ID, container.ExecAttachOptions{})
	if err != nil {
		session.Error("Failed to attach to exec: %v", err)
		return fmt.Errorf("failed to attach to exec: %w", err)
	}
	defer resp.Close()

	_, err = resp.Conn.Write([]byte("SET statement_timeout = 0;\n"))
	if err != nil {
		session.Error("Failed to write timeout setting: %v", err)
		return fmt.Errorf("failed to write timeout setting: %w", err)
	}

	_, err = io.Copy(resp.Conn, file)
	if err != nil {
		session.Error("Failed to copy backup data: %v", err)
		return fmt.Errorf("failed to copy backup data: %w", err)
	}

	resp.CloseWrite()

	execInspect, err := cli.ContainerExecInspect(ctx, execResp.ID)
	if err != nil {
		session.Error("Failed to inspect exec: %v", err)
		return fmt.Errorf("failed to inspect exec: %w", err)
	}

	if execInspect.ExitCode != 0 {
		session.Error("psql failed with exit code %d", execInspect.ExitCode)
		return fmt.Errorf("psql failed with exit code %d", execInspect.ExitCode)
	}

	return nil
}
