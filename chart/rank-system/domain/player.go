package domain

import (
	"time"
)

// Player 玩家实体
type Player struct {
	ID         int64
	Score      int64
	Rank       int
	UpdateTime time.Time
}

// NewPlayer 创建新玩家
func NewPlayer(id, score int64) *Player {
	return &Player{
		ID:         id,
		Score:      score,
		UpdateTime: time.Now(),
	}
}

// UpdateScore 更新分数
func (p *Player) UpdateScore(score int64) {
	p.Score = score
	p.UpdateTime = time.Now()
}

// PlayerList 玩家列表，用于排序
type PlayerList []*Player

func (p PlayerList) Len() int           { return len(p) }
func (p PlayerList) Less(i, j int) bool {
	if p[i].Score == p[j].Score {
		return p[i].UpdateTime.Before(p[j].UpdateTime)
	}
	return p[i].Score > p[j].Score
}
func (p PlayerList) Swap(i, j int) { p[i], p[j] = p[j], p[i] }