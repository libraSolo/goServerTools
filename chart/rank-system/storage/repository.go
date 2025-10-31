package storage

import "rank-system/domain"

// Repository 仓储接口
type Repository interface {
	Get(id string) (*domain.Leaderboard, error)
	Save(leaderboard *domain.Leaderboard) error
	Delete(id string) error
	Exists(id string) bool
}