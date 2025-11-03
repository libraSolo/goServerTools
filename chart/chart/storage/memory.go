package storage

import (
	"rank-system/domain"
	"sync"
)

// MemoryRepository 内存存储实现
type MemoryRepository struct {
	mu           sync.RWMutex
	leaderboards map[string]*domain.Leaderboard
}

// NewMemoryRepository 创建内存存储
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		leaderboards: make(map[string]*domain.Leaderboard),
	}
}

// SaveLeaderboard 保存排行榜
func (r *MemoryRepository) SaveLeaderboard(leaderboard *domain.Leaderboard) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.leaderboards[leaderboard.ID] = leaderboard
	return nil
}

// GetLeaderboard 获取排行榜
func (r *MemoryRepository) GetLeaderboard(id string) (*domain.Leaderboard, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	leaderboard, exists := r.leaderboards[id]
	if !exists {
		return nil, domain.ErrLeaderboardNotFound
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
	// 在内存存储中，玩家数据已经包含在Leaderboard中
	return nil
}

func (r *MemoryRepository) GetPlayer(leaderboardID string, playerID int64) (*domain.Player, error) {
	leaderboard, err := r.GetLeaderboard(leaderboardID)
	if err != nil {
		return nil, err
	}

	return leaderboard.GetPlayerRank(playerID)
}

func (r *MemoryRepository) RemovePlayer(leaderboardID string, playerID int64) error {
	// 在领域层实现
	return nil
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
