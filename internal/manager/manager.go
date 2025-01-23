package manager

import (
	"sync"

	"github.com/bytekai/docker-auto-backup/internal/logger"
	"github.com/bytekai/docker-auto-backup/internal/scheduler"
)

type Manager struct {
	schedulers map[string]scheduler.Scheduler
	mu         sync.RWMutex
	logger     *logger.Logger
}

func New(logger *logger.Logger) *Manager {
	return &Manager{
		schedulers: make(map[string]scheduler.Scheduler),
		logger:     logger,
	}
}

func (m *Manager) AddScheduler(containerID string, scheduler scheduler.Scheduler) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if existing, exists := m.schedulers[containerID]; exists {
		existing.Stop()
	}

	m.schedulers[containerID] = scheduler
}

func (m *Manager) RemoveScheduler(containerID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if scheduler, exists := m.schedulers[containerID]; exists {
		scheduler.Stop()
		delete(m.schedulers, containerID)
	}
}
