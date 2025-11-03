package storage

import (
	"rank-system/domain"
	"sync"
)

// MemoryRepository 内存仓储实现
type MemoryRepository struct {
	leaderboards map[string]*domain.Leaderboard
	mu           sync.RWMutex
}

// NewMemoryRepository 创建内存仓储
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		leaderboards: make(map[string]*domain.Leaderboard),
	}
}

// Get 获取排行榜
func (r *MemoryRepository) Get(id string) (*domain.Leaderboard, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	leaderboard, exists := r.leaderboards[id]
	if !exists {
		return nil, domain.ErrLeaderboardNotFound // 复用错误，实际应该定义新错误
	}

	// 返回副本以避免并发修改
	return r.cloneLeaderboard(leaderboard), nil
}

// Save 保存排行榜
func (r *MemoryRepository) Save(leaderboard *domain.Leaderboard) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.leaderboards[leaderboard.ID] = r.cloneLeaderboard(leaderboard)
	return nil
}

// Delete 删除排行榜
func (r *MemoryRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.leaderboards, id)
	return nil
}

// Exists 检查排行榜是否存在
func (r *MemoryRepository) Exists(id string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.leaderboards[id]
	return exists
}

// cloneLeaderboard 深拷贝排行榜
func (r *MemoryRepository) cloneLeaderboard(original *domain.Leaderboard) *domain.Leaderboard {
	cloned := domain.NewLeaderboard(original.ID, original.Name, original.Config)
	cloned.Version = original.Version
	cloned.CreatedAt = original.CreatedAt
	cloned.UpdatedAt = original.UpdatedAt

	// 拷贝玩家数据
	for _, p := range original.GetPlayers() {
		playerCopy := *p
		cloned.GetPlayers()[p.ID] = &playerCopy
	}

	// 拷贝排序列表
	sortedPlayers := make(domain.PlayerList, len(original.GetSortedPlayers()))
	copy(sortedPlayers, original.GetSortedPlayers())

	// 更新克隆对象的排序列表
	for i, p := range sortedPlayers {
		sortedPlayers[i] = cloned.GetPlayers()[p.ID]
	}
	cloned.SetSortedPlayers(sortedPlayers)
	cloned.SetIsDirty(original.IsDirty())

	return cloned
}