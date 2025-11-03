package storage

import "chart/domain"

// Repository 仓储接口
type Repository interface {
	// 排行榜管理
	SaveLeaderboard(leaderboard *domain.Leaderboard) error
	GetLeaderboard(id string) (*domain.Leaderboard, error)
	DeleteLeaderboard(id string) error
	ExistsLeaderboard(id string) bool

	// 玩家数据管理
	SavePlayer(leaderboardID string, player *domain.Player) error
	GetPlayer(leaderboardID string, playerID int64) (*domain.Player, error)
	RemovePlayer(leaderboardID string, playerID int64) error

	// 批量操作
	GetTopPlayers(leaderboardID string, limit int) ([]*domain.Player, error)
	GetPlayerCount(leaderboardID string) (int, error)
}
