package persistence

import (
	"encoding/gob"
	"leaderboard/internal/domain/model"
	"os"
)

// Snapshotter 负责创建和加载排行榜快照。
type Snapshotter struct {
	filePath string
}

// NewSnapshotter 创建一个新的 Snapshotter。
func NewSnapshotter(filePath string) *Snapshotter {
	return &Snapshotter{filePath: filePath}
}

// Save 创建排行榜的快照。
func (s *Snapshotter) Save(lb *model.Leaderboard) error {
	file, err := os.Create(s.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	return encoder.Encode(lb)
}

// Load 从快照文件中加载排行榜。
func (s *Snapshotter) Load() (*model.Leaderboard, error) {
	file, err := os.Open(s.filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)
	var lb model.Leaderboard
	if err := decoder.Decode(&lb); err != nil {
		return nil, err
	}
	return &lb, nil
}