package storage

import (
	"rank-system/domain"
	"sync"
)

// MultiBackendRepository 多后端存储
type MultiBackendRepository struct {
	backends []Repository
	strategy BackendStrategy
	mu       sync.RWMutex
}

// BackendStrategy 后端策略
type BackendStrategy interface {
	SelectBackend(leaderboardID string, operation string) Repository
}

// NewMultiBackendRepository 创建多后端存储
func NewMultiBackendRepository(backends ...Repository) *MultiBackendRepository {
	return &MultiBackendRepository{
		backends: backends,
		strategy: &DefaultStrategy{},
	}
}

// SaveLeaderboard 保存排行榜
func (m *MultiBackendRepository) SaveLeaderboard(leaderboard *domain.Leaderboard) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 写入所有后端
	var lastErr error
	for _, backend := range m.backends {
		if err := backend.SaveLeaderboard(leaderboard); err != nil {
			lastErr = err
			// 记录日志，但不中断其他后端写入
		}
	}
	return lastErr
}

// GetLeaderboard 获取排行榜
func (m *MultiBackendRepository) GetLeaderboard(id string) (*domain.Leaderboard, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 根据策略选择后端
	backend := m.strategy.SelectBackend(id, "read")
	return backend.GetLeaderboard(id)
}

// DefaultStrategy 默认策略
type DefaultStrategy struct{}

func (s *DefaultStrategy) SelectBackend(leaderboardID string, operation string) Repository {
	// 简单策略：第一个后端用于读，所有后端用于写
	// 实际可以根据负载、延迟等选择
	return globalBackends[0] // 假设全局变量
}
