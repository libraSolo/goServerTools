package application

import (
	"leaderboard/internal/domain/model"
	"leaderboard/internal/domain/repository"
)

// RankService 定义了排行榜应用服务。
type RankService interface {
	UpdateScore(playerID int64, score int64) error
	GetPlayerRank(playerID int64) (int64, error)
	GetTopN(n int) ([]*model.Player, error)
	GetNearbyRanks(playerID int64, count int) ([]*model.Player, error)
}

// rankServiceImpl 是 RankService 的实现。
type rankServiceImpl struct {
	leaderboardRepo repository.LeaderboardRepository
	leaderboard     *model.Leaderboard
}

// NewRankService 创建一个新的 RankService。
func NewRankService(leaderboard *model.Leaderboard, repo repository.LeaderboardRepository) (RankService, error) {
	return &rankServiceImpl{
		leaderboard:     leaderboard,
		leaderboardRepo: repo,
	}, nil
}

// UpdateScore 更新玩家的分数。
func (s *rankServiceImpl) UpdateScore(playerID int64, score int64) error {
	s.leaderboard.UpdateScore(playerID, score)
	return s.leaderboardRepo.LogUpdate(playerID, score)
}

// GetPlayerRank 获取玩家的排名。
func (s *rankServiceImpl) GetPlayerRank(playerID int64) (int64, error) {
	return s.leaderboard.GetPlayerRank(playerID)
}

// GetTopN 获取排名前 N 的玩家。
func (s *rankServiceImpl) GetTopN(n int) ([]*model.Player, error) {
	return s.leaderboard.GetTopN(n), nil
}

// GetNearbyRanks 获取玩家临近的排名。
func (s *rankServiceImpl) GetNearbyRanks(playerID int64, count int) ([]*model.Player, error) {
	return s.leaderboard.GetNearbyRanks(playerID, count)
}