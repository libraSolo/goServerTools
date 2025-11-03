package repository

import (
	"leaderboard/internal/domain/model"
)

// LeaderboardRepository 定义了排行榜的持久化接口。
type LeaderboardRepository interface {
	Save(*model.Leaderboard) error
	Load(id string) (*model.Leaderboard, error)
	LogUpdate(playerID int64, score int64) error
}