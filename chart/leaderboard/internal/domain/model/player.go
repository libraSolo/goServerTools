package model

import "time"

// Player 表示排行榜中的一个玩家。
type Player struct {
    ID        int64     `json:"id"`
    Score     int64     `json:"score"`
    UpdatedAt time.Time `json:"updated_at"`
}

// NewPlayer 创建一个新玩家。
func NewPlayer(id int64, score int64) *Player {
	return &Player{
		ID:        id,
		Score:     score,
		UpdatedAt: time.Now(),
	}
}