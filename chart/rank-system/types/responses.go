package types

import (
	"rank-system/domain"
	"time"
)

// LeaderboardResponse 定义了查询排行榜信息时的响应结构。
type LeaderboardResponse struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Type        string          `json:"type"`
	PlayerCount int             `json:"player_count"`
	TopScore    int64           `json:"top_score"`
	Players     []*domain.Player `json:"players,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// PlayerRankResponse 定义了查询玩家排名时的响应结构。
type PlayerRankResponse struct {
	*domain.Player
	TotalPlayers int     `json:"total_players"`
	Percentile   float64 `json:"percentile"` // 百分比排名
}

// LeaderboardStatsResponse 定义了查询排行榜统计信息时的响应结构。
type LeaderboardStatsResponse struct {
	LeaderboardID string    `json:"leaderboard_id"`
	TotalPlayers  int       `json:"total_players"`
	AverageScore  float64   `json:"average_score"`
	MedianScore   int64     `json:"median_score"`
	TopScore      int64     `json:"top_score"`
	UpdateTime    time.Time `json:"update_time"`
}

// BatchResult 定义了批量操作的结果，包括成功、失败和错误详情。
type BatchResult struct {
	Total   int           `json:"total"`
	Success int           `json:"success"`
	Failed  int           `json:y`
	Errors  []*BatchError `json:"errors,omitempty"`
}

// BatchError 定义了批量操作中单个失败的错误详情。
type BatchError struct {
	PlayerID int64  `json:"player_id"`
	Error    string `json:"error"`
}