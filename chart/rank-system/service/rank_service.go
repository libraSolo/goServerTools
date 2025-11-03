package service

import (
	"rank-system/domain"
	"rank-system/storage"
	"rank-system/types"
	"sync"
)

// RankService 排名应用服务
type RankService struct {
	repo storage.Repository
}

// NewRankService 创建排名服务
func NewRankService(repo storage.Repository) *RankService {
	return &RankService{
		repo: repo,
	}
}

// BatchUpdateScore 批量更新玩家分数
func (s *RankService) BatchUpdateScore(req *types.BatchUpdateScoreRequest) (*types.BatchResult, error) {
	leaderboard, err := s.repo.Get(req.LeaderboardID)
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	results := &types.BatchResult{}

	for _, update := range req.Updates {
		wg.Add(1)
		go func(u *types.ScoreUpdate) {
			defer wg.Done()
			leaderboard.UpdatePlayerScore(u.PlayerID, u.Score)
		}(update)
	}

	wg.Wait()

	if err := s.repo.Save(leaderboard); err != nil {
		return nil, err
	}

	results.Success = len(req.Updates)
	return results, nil
}

// GetPlayerRank 获取玩家排名
func (s *RankService) GetPlayerRank(req *types.QueryLeaderboardRequest) (*types.PlayerRankResponse, error) {
	leaderboard, err := s.repo.Get(req.LeaderboardID)
	if err != nil {
		return nil, err
	}

	player, err := leaderboard.GetPlayerRank(req.PlayerID)
	if err != nil {
		return nil, err
	}

	return &types.PlayerRankResponse{Player: player}, nil
}

// GetNearbyRanks 获取临近排名
func (s *RankService) GetNearbyRanks(req *types.QueryLeaderboardRequest) (*types.LeaderboardResponse, error) {
	leaderboard, err := s.repo.Get(req.LeaderboardID)
	if err != nil {
		return nil, err
	}

	nearbyRanks, err := leaderboard.GetNearbyRanks(req.PlayerID, req.PageSize)
	if err != nil {
		return nil, err
	}

	return &types.LeaderboardResponse{Players: nearbyRanks}, nil
}

// GetTopRanks 获取前N名
func (s *RankService) GetTopRanks(req *types.QueryLeaderboardRequest) (*types.LeaderboardResponse, error) {
	leaderboard, err := s.repo.Get(req.LeaderboardID)
	if err != nil {
		return nil, err
	}

	topRanks := leaderboard.GetTopRanks(req.PageSize)
	return &types.LeaderboardResponse{Players: topRanks}, nil
}

// CreateLeaderboard 创建排行榜
func (s *RankService) CreateLeaderboard(req *types.CreateLeaderboardRequest) error {
	config := domain.NewRankConfig(
		req.TotalPlayers,
		req.RewardRatio,
		req.MinReward,
		req.MaxReward,
	)

	leaderboard := domain.NewLeaderboard(req.ID, req.Name, config)
	return s.repo.Save(leaderboard)
}