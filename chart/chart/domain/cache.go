// RankCache 前 N 名结果的轻量级缓存
//
// 设计要点：
// - 以 limit 作为键缓存不同维度的 TopN 结果；
// - 使用短 TTL（duration）权衡实时性与性能；
// - 读写分离锁：Get 使用 RLock，Set/Invalidate 使用 Lock，避免数据竞争；
// - Get 返回副本以避免外部修改导致共享数据不一致。
package domain

import (
	"sync"
	"time"
)

// RankCache 排名缓存
type RankCache struct {
	mu        sync.RWMutex
	topRanks  map[int][]*Player // limit -> players
	cacheTime map[int]time.Time // limit -> cache time
	duration  time.Duration
}

// NewRankCache 创建排名缓存
func NewRankCache(duration time.Duration) *RankCache {
	return &RankCache{
		topRanks:  make(map[int][]*Player),
		cacheTime: make(map[int]time.Time),
		duration:  duration,
	}
}

// GetTopRanks 获取缓存的前N名
func (c *RankCache) GetTopRanks(limit int) []*Player {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if players, exists := c.topRanks[limit]; exists {
		if time.Since(c.cacheTime[limit]) < c.duration {
			// 返回副本避免数据竞争
			result := make([]*Player, len(players))
			copy(result, players)
			return result
		}
	}
	return nil
}

// SetTopRanks 设置前N名缓存
func (c *RankCache) SetTopRanks(limit int, players []*Player) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.topRanks[limit] = players
	c.cacheTime[limit] = time.Now()
}

// Invalidate 使缓存失效
func (c *RankCache) Invalidate() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.topRanks = make(map[int][]*Player)
	c.cacheTime = make(map[int]time.Time)
}
