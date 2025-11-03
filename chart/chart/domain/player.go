// 玩家实体
package domain

import "time"

// Player 玩家实体
type Player struct {
	ID         int64     `json:"id"`          // 玩家ID
	Score      int64     `json:"score"`       // 玩家分数
	Rank       int       `json:"rank"`        // 玩家排名
	UpdateTime time.Time `json:"update_time"` // 玩家更新时间
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
