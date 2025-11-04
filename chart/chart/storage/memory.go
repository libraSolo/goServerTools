package storage

import (
    "chart/domain"
    "errors"
    "sync"
)

// MemoryRepository 内存存储实现
type MemoryRepository struct {
    mu           sync.RWMutex
    leaderboards map[string]*domain.HybridLeaderboard
}

// NewMemoryRepository 创建内存存储
func NewMemoryRepository() *MemoryRepository {
    return &MemoryRepository{
        leaderboards: make(map[string]*domain.HybridLeaderboard),
    }
}

// SaveLeaderboard 保存排行榜
func (r *MemoryRepository) SaveLeaderboard(leaderboard *domain.HybridLeaderboard) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    r.leaderboards[leaderboard.ID] = leaderboard
    return nil
}

// GetLeaderboard 获取排行榜
func (r *MemoryRepository) GetLeaderboard(id string) (*domain.HybridLeaderboard, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    leaderboard, exists := r.leaderboards[id]
    if !exists {
        return nil, errors.New("leaderboard not found")
    }

    return leaderboard, nil
}

// DeleteLeaderboard 删除排行榜
func (r *MemoryRepository) DeleteLeaderboard(id string) error {
    r.mu.Lock()
    defer r.mu.Unlock()

	delete(r.leaderboards, id)
	return nil
}

// ExistsLeaderboard 检查排行榜是否存在
func (r *MemoryRepository) ExistsLeaderboard(id string) bool {
    r.mu.RLock()
    defer r.mu.RUnlock()

	_, exists := r.leaderboards[id]
	return exists
}

// 以下方法在内存存储中直接通过领域对象操作，无需单独实现
func (r *MemoryRepository) SavePlayer(leaderboardID string, player *domain.Player) error {
    r.mu.RLock()
    lb, ok := r.leaderboards[leaderboardID]
    r.mu.RUnlock()
    if !ok {
        return errors.New("leaderboard not found")
    }
    // 将玩家当前分数写入排行榜（作为一次更新）
    return lb.UpdateScore(player.ID, player.Score)
}

func (r *MemoryRepository) GetPlayer(leaderboardID string, playerID int64) (*domain.Player, error) {
    // 本地领域未导出直接按 ID 获取玩家实体的 API，
    // 这里返回 nil 表示未实现具体读路径（当前接口层未使用）。
    return nil, errors.New("GetPlayer not implemented in memory repository")
}

func (r *MemoryRepository) RemovePlayer(leaderboardID string, playerID int64) error {
    // 领域暂未提供移除 API，此处留空实现
    return errors.New("RemovePlayer not implemented")
}

func (r *MemoryRepository) GetTopPlayers(leaderboardID string, limit int) ([]*domain.Player, error) {
    leaderboard, err := r.GetLeaderboard(leaderboardID)
    if err != nil {
        return nil, err
    }

    return leaderboard.GetTopRanks(limit), nil
}

func (r *MemoryRepository) GetPlayerCount(leaderboardID string) (int, error) {
    leaderboard, err := r.GetLeaderboard(leaderboardID)
    if err != nil {
        return 0, err
    }

    return leaderboard.GetPlayerCount(), nil
}
