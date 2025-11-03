package types

import "time"

// CreateLeaderboardRequest 定义了创建排行榜时所需的请求体结构。
type CreateLeaderboardRequest struct {
	ID           string  `json:"id" binding:"required,alphanum"`
	Name         string  `json:"name" binding:"required"`
	Type         string  `json:"type" binding:"required"`
	TotalPlayers int     `json:"total_players" binding:"min=1"`
	RewardRatio  float64 `json:"reward_ratio" binding:"min=0,max=1"`
	MinReward    int     `json:"min_reward" binding:"min=1"`
	MaxReward    int     `json:"max_reward" binding:"min=1"`
}

// BatchUpdateScoreRequest 定义了批量更新分数时所需的请求体结构。
type BatchUpdateScoreRequest struct {
	LeaderboardID string         `json:"leaderboard_id" binding:"required"`
	Updates       []*ScoreUpdate `json:"updates" binding:"required,dive"`
}

// ScoreUpdate 定义了单个分数更新的数据结构。
type ScoreUpdate struct {
	PlayerID int64 `json:"player_id" binding:"required"`
	Score    int64 `json:"score" binding:"required"`
}

// QueryLeaderboardRequest 定义了查询排行榜时的请求参数结构。
type QueryLeaderboardRequest struct {
	PageRequest
	LeaderboardID string    `json:"leaderboard_id" form:"leaderboard_id"`
	PlayerID      int64     `json:"player_id" form:"player_id"`
	PageSize      int       `json:"page_size" form:"page_size"`
	Type          string    `json:"type" form:"type"`
	Status        string    `json:"status" form:"status"`
	Keywords      string    `json:"keywords" form:"keywords"`
	FromTime      time.Time `json:"from_time" form:"from_time"`
	ToTime        time.Time `json:"to_time" form:"to_time"`
}