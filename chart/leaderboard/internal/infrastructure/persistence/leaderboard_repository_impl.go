package persistence

import (
	"leaderboard/internal/domain/model"
	"leaderboard/internal/domain/repository"
	"os"
)

// leaderboardRepositoryImpl 是 LeaderboardRepository 的实现。
type leaderboardRepositoryImpl struct {
	snapshotter *Snapshotter
	aofLogger   *AOFLogger
}

// NewLeaderboardRepository 创建一个新的 leaderboardRepositoryImpl。
func NewLeaderboardRepository(dataDir string, id string) (*model.Leaderboard, repository.LeaderboardRepository, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, nil, err
	}

	snapshotter := NewSnapshotter(dataDir + "/snapshot.gob")
	aofLogger, err := NewAOFLogger(dataDir + "/aof.log")
	if err != nil {
		return nil, nil, err
	}

	repo := &leaderboardRepositoryImpl{
		snapshotter: snapshotter,
		aofLogger:   aofLogger,
	}

	lb, err := repo.Load(id)
	if err != nil {
		return nil, nil, err
	}

	return lb, repo, nil
}

// Save 保存排行榜快照。
func (r *leaderboardRepositoryImpl) Save(lb *model.Leaderboard) error {
	return r.snapshotter.Save(lb)
}

// Load 加载排行榜。
func (r *leaderboardRepositoryImpl) Load(id string) (*model.Leaderboard, error) {
	lb, err := r.snapshotter.Load()
	if err != nil {
		// 如果快照不存在，则创建一个新的排行榜
		if os.IsNotExist(err) {
			return model.NewLeaderboard(id, "default"), nil
		}
		return nil, err
	}

	// 回放 AOF 日志
	if err := r.aofLogger.Replay(lb); err != nil {
		return nil, err
	}

	return lb, nil
}

// LogUpdate 记录分数更新。
func (r *leaderboardRepositoryImpl) LogUpdate(playerID int64, score int64) error {
	return r.aofLogger.LogUpdate(playerID, score)
}