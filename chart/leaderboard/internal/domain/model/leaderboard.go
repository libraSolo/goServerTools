package model

import (
	"errors"
	"sync"
)

var (
	ErrPlayerNotFound = errors.New("player not found")
)

// Leaderboard 是排行榜的聚合根。
type Leaderboard struct {
	ID      string
	Name    string
	players map[int64]*Node
	sl      *SkipList
	mu      sync.RWMutex
}

// NewLeaderboard 创建一个新的排行榜。
func NewLeaderboard(id, name string) *Leaderboard {
	return &Leaderboard{
		ID:      id,
		Name:    name,
		players: make(map[int64]*Node),
		sl:      NewSkipList(),
	}
}

// UpdateScore 更新玩家的分数。
func (l *Leaderboard) UpdateScore(playerID int64, score int64) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if node, ok := l.players[playerID]; ok {
		// 如果分数没有变化，则不更新
		if node.Player.Score == score {
			return
		}
		// 从跳表中删除旧节点
		l.sl.Delete(node.Player.Score, node.Player.ID)
	}

	player := NewPlayer(playerID, score)
	node := l.sl.Insert(player)
	l.players[playerID] = node
}

// GetPlayerRank 获取玩家的排名。
func (l *Leaderboard) GetPlayerRank(playerID int64) (int64, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if node, ok := l.players[playerID]; ok {
		rank := l.sl.GetRank(node.Player.Score, node.Player.ID)
		return rank, nil
	}

	return 0, ErrPlayerNotFound
}

// GetTopN 获取排名前 N 的玩家。
func (l *Leaderboard) GetTopN(n int) []*Player {
	l.mu.RLock()
	defer l.mu.RUnlock()

	players := make([]*Player, 0, n)
	node := l.sl.header.level[0].forward
	for i := 0; i < n && node != nil; i++ {
		players = append(players, node.Player)
		node = node.level[0].forward
	}
	return players
}

// GetNearbyRanks 获取玩家临近的排名。
func (l *Leaderboard) GetNearbyRanks(playerID int64, count int) ([]*Player, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if node, ok := l.players[playerID]; ok {
		rank := l.sl.GetRank(node.Player.Score, node.Player.ID)
		startRank := rank - int64(count/2)
		if startRank < 1 {
			startRank = 1
		}

		players := make([]*Player, 0, count)
		startNode := l.sl.GetElementByRank(startRank)
		if startNode == nil {
			return players, nil
		}
		for i := 0; i < count && startNode != nil; i++ {
			players = append(players, startNode.Player)
			startNode = startNode.level[0].forward
		}
		return players, nil
	}

	return nil, ErrPlayerNotFound
}