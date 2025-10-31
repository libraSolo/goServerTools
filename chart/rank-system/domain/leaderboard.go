package domain

import (
	"errors"
	"sort"
	"time"
)

// Leaderboard 排行榜聚合根
type Leaderboard struct {
	ID          string
	Name        string
	Config      *RankConfig
	players     map[int64]*Player // 玩家数据
	sorted      PlayerList        // 排序后的玩家列表
	isDirty     bool              // 标记是否需要重新排序
	Version     int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// RankConfig 排行榜配置
type RankConfig struct {
	TotalPlayers int
	RewardRatio  float64
	MinReward    int
	MaxReward    int
}

// NewRankConfig 创建排行榜配置
func NewRankConfig(totalPlayers int, rewardRatio float64, minReward, maxReward int) *RankConfig {
	return &RankConfig{
		TotalPlayers: totalPlayers,
		RewardRatio:  rewardRatio,
		MinReward:    minReward,
		MaxReward:    maxReward,
	}
}

// NewLeaderboard 创建新排行榜
func NewLeaderboard(id, name string, config *RankConfig) *Leaderboard {
	return &Leaderboard{
		ID:        id,
		Name:      name,
		Config:    config,
		players:   make(map[int64]*Player),
		sorted:    make(PlayerList, 0),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// UpdatePlayerScore 更新玩家分数
func (l *Leaderboard) UpdatePlayerScore(playerID, score int64) {
	player, exists := l.players[playerID]
	if exists {
		player.UpdateScore(score)
	} else {
		player = NewPlayer(playerID, score)
		l.players[playerID] = player
		l.sorted = append(l.sorted, player)
	}

	l.isDirty = true
	l.UpdatedAt = time.Now()
	l.Version++
}

// GetPlayerRank 获取玩家排名
func (l *Leaderboard) GetPlayerRank(playerID int64) (*Player, error) {
	player, exists := l.players[playerID]
	if !exists {
		return nil, ErrPlayerNotFound
	}

	// 确保数据已排序
	l.ensureSorted()

	return player, nil
}

// GetNearbyRanks 获取临近排名
func (l *Leaderboard) GetNearbyRanks(playerID int64, rangeSize int) ([]*Player, error) {
	player, err := l.GetPlayerRank(playerID)
	if err != nil {
		return nil, err
	}

	l.ensureSorted()

	start := max(0, player.Rank-1-rangeSize)
	end := min(len(l.sorted), player.Rank-1+rangeSize+1)

	result := make([]*Player, end-start)
	copy(result, l.sorted[start:end])

	return result, nil
}

// GetTopRanks 获取前N名
func (l *Leaderboard) GetTopRanks(count int) []*Player {
	l.ensureSorted()

	numPlayers := len(l.sorted)

	if count <= 0 {
		// If count is not specified, use the reward count as default
		rewardCount := l.calculateRewardCount()
		count = min(rewardCount, numPlayers)
	} else {
		count = min(count, numPlayers)
	}

	result := make([]*Player, count)
	copy(result, l.sorted[:count])

	return result
}

// GetPlayerCount 获取玩家数量
func (l *Leaderboard) GetPlayerCount() int {
	return len(l.players)
}

// GetPlayers 获取所有玩家
func (l *Leaderboard) GetPlayers() map[int64]*Player {
	return l.players
}

// GetSortedPlayers 获取排序后的玩家列表
func (l *Leaderboard) GetSortedPlayers() PlayerList {
	return l.sorted
}

// SetSortedPlayers 设置排序后的玩家列表
func (l *Leaderboard) SetSortedPlayers(players PlayerList) {
	l.sorted = players
}

// IsDirty 获取是否需要重新排序
func (l *Leaderboard) IsDirty() bool {
	return l.isDirty
}

// SetIsDirty 设置是否需要重新排序
func (l *Leaderboard) SetIsDirty(isDirty bool) {
	l.isDirty = isDirty
}

// ensureSorted 确保玩家列表已排序
func (l *Leaderboard) ensureSorted() {
	if !l.isDirty {
		return
	}

	sort.Sort(l.sorted)

	// 更新排名
	for i, player := range l.sorted {
		player.Rank = i + 1
	}

	l.isDirty = false
}

// calculateRewardCount 计算奖励人数
func (l *Leaderboard) calculateRewardCount() int {
	total := len(l.players)
	rewardCount := int(float64(total) * l.Config.RewardRatio)
	rewardCount = max(rewardCount, l.Config.MinReward)
	rewardCount = min(rewardCount, l.Config.MaxReward)
	return rewardCount
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

// 错误定义
var (
	ErrPlayerNotFound      = errors.New("player not found")
	ErrLeaderboardNotFound = errors.New("leaderboard not found")
)