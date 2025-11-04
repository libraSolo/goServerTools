package domain

import (
	"container/heap"
	"errors"
	"sync"
	"time"
)

// RankConfig 排行榜配置
type RankConfig struct {
	TotalPlayers int     `json:"total_players"` // 总玩家数
	RewardRatio  float64 `json:"reward_ratio"`  // 奖励比例
	MinReward    int     `json:"min_reward"`    // 最小奖励
	MaxReward    int     `json:"max_reward"`    // 最大奖励
}

type ScoreUpdate struct {
	PlayerID int64 `json:"player_id" binding:"required"` // 玩家ID
	Score    int64 `json:"score" binding:"required"`     // 玩家分数
}

// HybridLeaderboard 混合策略排行榜（跳表 + 分段）
type HybridLeaderboard struct {
	mu     sync.RWMutex
	ID     string
	Name   string
	Config *RankConfig

	// 核心数据结构
	skipList  *SkipList         // 跳表 - 用于精确排名计算
	topK      int               // 维护前K名 没有降分淘汰，其topk的容量比返回大给定兼容即可
	topHeap   *TopPlayersHeap   // 前K名最小堆 - 用于快速获取前N名
	playerMap map[int64]*Player // 所有玩家数据 - O(1)查找
	topMap    map[int64]*Player // 前K名玩家快速查找

	// 性能优化
	batchUpdates chan *ScoreUpdate // 批量更新通道
	cache        *RankCache        // 排名缓存
	version      int64             // 版本控制
}

// NewHybridLeaderboard 创建混合策略排行榜
func NewHybridLeaderboard(id, name string, config *RankConfig) *HybridLeaderboard {
	lb := &HybridLeaderboard{
		ID:           id,
		Name:         name,
		Config:       config,
		skipList:     NewSkipList(),
		topK:         1000,
		topHeap:      &TopPlayersHeap{},
		playerMap:    make(map[int64]*Player),
		topMap:       make(map[int64]*Player),
		batchUpdates: make(chan *ScoreUpdate, 10000),
		cache:        NewRankCache(2 * time.Second),
	}

	heap.Init(lb.topHeap)
	go lb.processBatchUpdates()

	return lb
}

// UpdateScore 更新玩家分数 - O(log n)
func (lb *HybridLeaderboard) UpdateScore(playerID, score int64) error {
	update := &ScoreUpdate{
		PlayerID: playerID,
		Score:    score,
	}

	select {
	case lb.batchUpdates <- update:
		return nil
	default:
		return lb.syncUpdateScore(playerID, score)
	}
}

// processBatchUpdates 处理批量更新
func (lb *HybridLeaderboard) processBatchUpdates() {
	batch := make([]*ScoreUpdate, 0, 100)
	ticker := time.NewTicker(50 * time.Millisecond) // 更快的批处理
	defer ticker.Stop()

	// 单线程处理批量更新，不存在竞态
	for {
		select {
		case update, ok := <-lb.batchUpdates:
			if !ok {
				if len(batch) > 0 {
					lb.processBatch(batch)
				}
				return
			}
			batch = append(batch, update)
			if len(batch) >= 100 {
				lb.processBatch(batch)
				batch = batch[:0]
			}
		case <-ticker.C:
			if len(batch) > 0 {
				lb.processBatch(batch)
				batch = batch[:0]
			}
		}
	}
}

// Close 关闭排行榜 - 释放资源
func (lb *HybridLeaderboard) Close() {
	close(lb.batchUpdates)
}

// processBatch 批量处理更新
func (lb *HybridLeaderboard) processBatch(updates []*ScoreUpdate) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	for _, update := range updates {
		lb.applySingleUpdate(update.PlayerID, update.Score)
	}

	lb.version++
	lb.cache.Invalidate()
}

// applySingleUpdate 应用单个更新
func (lb *HybridLeaderboard) applySingleUpdate(playerID, score int64) {
	player, exists := lb.playerMap[playerID]

	if !exists {
		// 新玩家
		player = NewPlayer(playerID, score)
		lb.playerMap[playerID] = player
		lb.skipList.Insert(player)

		// 检查是否应该进入前K名
		if lb.shouldPromoteToTop(score) {
			lb.promoteToTop(player)
		}
	} else {
		// 更新现有玩家
		//oldScore := player.Score
		lb.skipList.UpdateScore(player, score)

		// 更新前K名逻辑
		if _, inTop := lb.topMap[playerID]; inTop {
			lb.adjustTopPlayer(player)
		} else if lb.shouldPromoteToTop(score) {
			lb.promoteToTop(player)
		}
	}
}

// shouldPromoteToTop 判断是否应该进入前K名
func (lb *HybridLeaderboard) shouldPromoteToTop(score int64) bool {
	if lb.topHeap.Len() < lb.topK {
		return true
	}
	return score > (*lb.topHeap)[0].Score
}

// promoteToTop 提升玩家到前K名
func (lb *HybridLeaderboard) promoteToTop(player *Player) {
	if lb.topHeap.Len() >= lb.topK {
		// 移除最低分玩家
		removed := heap.Pop(lb.topHeap).(*Player)
		delete(lb.topMap, removed.ID)
	}

	heap.Push(lb.topHeap, player)
	lb.topMap[player.ID] = player
}

// adjustTopPlayer 调整前K名玩家位置
func (lb *HybridLeaderboard) adjustTopPlayer(player *Player) {

	for index, p := range *lb.topHeap {
		if player.ID == p.ID {
			heap.Fix(lb.topHeap, index)
		}
	}
}

// GetPlayerRank 获取玩家排名 - O(log n)
func (lb *HybridLeaderboard) GetPlayerRank(playerID int64) (int, error) {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	player, exists := lb.playerMap[playerID]
	if !exists {
		return 0, errors.New("player not found")
	}

	// 使用跳表基于排序键获取精确排名
	rank, found := lb.skipList.GetRankByPlayer(player)
	if !found {
		return 0, errors.New("player not found")
	}

	return rank, nil
}

// GetTopRanks 获取前N名 - O(1) 从堆中获取
func (lb *HybridLeaderboard) GetTopRanks(limit int) []*Player {
	// 尝试从缓存获取
	if cached := lb.cache.GetTopRanks(limit); cached != nil {
		return cached
	}

	return lb.refreshTopRanks(limit)
}

// refreshTopRanks 刷新前N名缓存
func (lb *HybridLeaderboard) refreshTopRanks(limit int) []*Player {
    lb.mu.RLock()
    defer lb.mu.RUnlock()

    // 直接使用跳表获取前 N 名，保证顺序正确
    if limit > lb.skipList.Length() {
        limit = lb.skipList.Length()
    }
    original := lb.skipList.GetRange(1, limit)
    // 返回副本并填充 Rank，避免修改共享实体导致竞态
    ranked := make([]*Player, len(original))
    for i, p := range original {
        ranked[i] = &Player{
            ID:         p.ID,
            Score:      p.Score,
            Rank:       i + 1,
            UpdateTime: p.UpdateTime,
        }
    }

    lb.cache.SetTopRanks(limit, ranked)
    return ranked
}

// GetNearbyRanks 获取临近排名 - O(log n + k)
func (lb *HybridLeaderboard) GetNearbyRanks(playerID int64, rangeSize int) ([]*Player, error) {
    lb.mu.RLock()
    defer lb.mu.RUnlock()

    rank, err := lb.GetPlayerRank(playerID)
    if err != nil {
        return nil, err
    }

    start := max(1, rank-rangeSize)
    end := min(lb.skipList.Length(), rank+rangeSize)
    original := lb.skipList.GetRange(start, end)
    // 返回副本并填充 Rank，避免修改共享实体导致竞态
    ranked := make([]*Player, len(original))
    for i, p := range original {
        ranked[i] = &Player{
            ID:         p.ID,
            Score:      p.Score,
            Rank:       start + i,
            UpdateTime: p.UpdateTime,
        }
    }
    return ranked, nil
}

// GetPlayerCount 获取玩家数量 - O(1)
func (lb *HybridLeaderboard) GetPlayerCount() int {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	return len(lb.playerMap)
}

// syncUpdateScore 同步更新分数
func (lb *HybridLeaderboard) syncUpdateScore(playerID, score int64) error {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	lb.applySingleUpdate(playerID, score)
	lb.version++
	lb.cache.Invalidate()

	return nil
}

// 工具函数
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
