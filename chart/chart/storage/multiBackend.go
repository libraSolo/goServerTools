package storage

import (
    "chart/domain"
    "errors"
    "sync"
)

// MultiBackendRepository 多后端存储
type MultiBackendRepository struct {
    backends []Repository
    mu       sync.RWMutex
}

// NewMultiBackendRepository 创建多后端存储
func NewMultiBackendRepository(backends ...Repository) *MultiBackendRepository {
    return &MultiBackendRepository{
        backends: backends,
    }
}

// SaveLeaderboard 保存排行榜
func (m *MultiBackendRepository) SaveLeaderboard(leaderboard *domain.HybridLeaderboard) error {
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
func (m *MultiBackendRepository) GetLeaderboard(id string) (*domain.HybridLeaderboard, error) {
    m.mu.RLock()
    defer m.mu.RUnlock()

    if len(m.backends) == 0 {
        return nil, errors.New("no backend available")
    }
    // 简单策略：选择第一个后端进行读取
    return m.backends[0].GetLeaderboard(id)
}
