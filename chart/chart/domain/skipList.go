package domain

import (
	"math/rand"
	"sync"
	"time"
)

// SkipListNode 跳表节点
type SkipListNode struct {
	Player   *Player
	Backward *SkipListNode
	Level    []SkipListLevel
}

// SkipListLevel 跳表层级
type SkipListLevel struct {
	Forward *SkipListNode
	Span    int
}

// SkipList 跳表
type SkipList struct {
	header *SkipListNode
	tail   *SkipListNode
	length int
	level  int
	mu     sync.RWMutex
}

const (
	maxSkipListLevel = 32
	skipListP        = 0.25
)

// NewSkipList 创建跳表
func NewSkipList() *SkipList {
	sl := &SkipList{
		level: 1,
		header: &SkipListNode{
			Level: make([]SkipListLevel, maxSkipListLevel),
		},
	}
	return sl
}

// randomLevel 随机生成层级
func (sl *SkipList) randomLevel() int {
	level := 1
	for rand.Float32() < skipListP && level < maxSkipListLevel {
		level++
	}
	return level
}

// Delete 删除节点
func (sl *SkipList) Delete(playerID int64) bool {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	update := make([]*SkipListNode, maxSkipListLevel)
	x := sl.header

	// 查找节点
	for i := sl.level - 1; i >= 0; i-- {
		for x.Level[i].Forward != nil &&
			(x.Level[i].Forward.Player.Score > sl.header.Player.Score ||
				(x.Level[i].Forward.Player.Score == sl.header.Player.Score &&
					x.Level[i].Forward.Player.ID != playerID)) {
			x = x.Level[i].Forward
		}
		update[i] = x
	}

	x = x.Level[0].Forward
	if x != nil && x.Player.ID == playerID {
		sl.deleteNode(x.Player.ID)
		return true
	}
	return false
}

// GetRange 获取排名范围内的玩家
func (sl *SkipList) GetRange(start, end int) []*Player {
	sl.mu.RLock()
	defer sl.mu.RUnlock()

	if start < 1 {
		start = 1
	}
	if end > sl.length {
		end = sl.length
	}
	if start > end {
		return nil
	}

	result := make([]*Player, 0, end-start+1)
	rank := 0
	x := sl.header

	// 移动到起始位置
	for i := sl.level - 1; i >= 0; i-- {
		for x.Level[i].Forward != nil && (rank+x.Level[i].Span) <= start {
			rank += x.Level[i].Span
			x = x.Level[i].Forward
		}
	}

	rank++ // 移动到第一个节点
	x = x.Level[0].Forward

	// 遍历范围内的节点
	for x != nil && rank <= end {
		result = append(result, x.Player)
		x = x.Level[0].Forward
		rank++
	}

	return result
}

// Length 获取跳表长度
func (sl *SkipList) Length() int {
	sl.mu.RLock()
	defer sl.mu.RUnlock()
	return sl.length
}

// 比较函数 - 统一分数比较逻辑
func comparePlayers(p1, p2 *Player) int {
	if p1.Score > p2.Score {
		return 1
	}
	if p1.Score < p2.Score {
		return -1
	}
	// 分数相同时，按更新时间排序（先更新的排前面）
	if p1.UpdateTime.Before(p2.UpdateTime) {
		return 1
	}
	if p1.UpdateTime.After(p2.UpdateTime) {
		return -1
	}
	// 更新时间也相同，按ID排序
	if p1.ID < p2.ID {
		return 1
	}
	if p1.ID > p2.ID {
		return -1
	}
	return 0
}

// Insert 插入节点（优化版）
func (sl *SkipList) Insert(player *Player) {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	update := make([]*SkipListNode, maxSkipListLevel)
	rank := make([]int, maxSkipListLevel)
	x := sl.header

	// 从最高层开始查找插入位置
	for i := sl.level - 1; i >= 0; i-- {
		if i == sl.level-1 {
			rank[i] = 0
		} else {
			rank[i] = rank[i+1]
		}

		for x.Level[i].Forward != nil &&
			comparePlayers(x.Level[i].Forward.Player, player) > 0 {
			rank[i] += x.Level[i].Span
			x = x.Level[i].Forward
		}
		update[i] = x
	}

	// 随机生成层级
	level := sl.randomLevel()
	if level > sl.level {
		for i := sl.level; i < level; i++ {
			rank[i] = 0
			update[i] = sl.header
			update[i].Level[i].Span = sl.length
		}
		sl.level = level
	}

	// 创建新节点（不再存储冗余分数）
	x = &SkipListNode{
		Player: player,
		Level:  make([]SkipListLevel, level),
	}

	// 更新指针
	for i := 0; i < level; i++ {
		x.Level[i].Forward = update[i].Level[i].Forward
		update[i].Level[i].Forward = x
		x.Level[i].Span = update[i].Level[i].Span - (rank[0] - rank[i])
		update[i].Level[i].Span = (rank[0] - rank[i]) + 1
	}

	// 更新高层级的span
	for i := level; i < sl.level; i++ {
		update[i].Level[i].Span++
	}

	// 更新后退指针
	if update[0] == sl.header {
		x.Backward = nil
	} else {
		x.Backward = update[0]
	}
	if x.Level[0].Forward != nil {
		x.Level[0].Forward.Backward = x
	} else {
		sl.tail = x
	}

	sl.length++
}

// GetRank 获取排名（优化版）
func (sl *SkipList) GetRank(playerID int64) (int, bool) {
	sl.mu.RLock()
	defer sl.mu.RUnlock()

	rank := 0
	x := sl.header

	// 从最高层开始查找
	for i := sl.level - 1; i >= 0; i-- {
		for x.Level[i].Forward != nil &&
			x.Level[i].Forward.Player.ID != playerID {
			rank += x.Level[i].Span
			x = x.Level[i].Forward
		}
	}

	x = x.Level[0].Forward
	if x != nil && x.Player.ID == playerID {
		return rank + 1, true
	}
	return 0, false
}

// UpdateScore 更新分数（需要删除再插入）
func (sl *SkipList) UpdateScore(player *Player, newScore int64) {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	// 先删除旧节点
	if sl.deleteNode(player.ID) {
		// 更新玩家分数
		player.Score = newScore
		player.UpdateTime = time.Now()
		// 重新插入
		sl.insertNode(player)
	}
}

// deleteNode 内部删除节点方法
func (sl *SkipList) deleteNode(playerID int64) bool {
	update := make([]*SkipListNode, maxSkipListLevel)
	x := sl.header

	// 查找节点
	for i := sl.level - 1; i >= 0; i-- {
		for x.Level[i].Forward != nil &&
			x.Level[i].Forward.Player.ID != playerID {
			x = x.Level[i].Forward
		}
		update[i] = x
	}

	x = x.Level[0].Forward
	if x != nil && x.Player.ID == playerID {
		// 删除节点逻辑...
		for i := 0; i < sl.level; i++ {
			if update[i].Level[i].Forward == x {
				update[i].Level[i].Span += x.Level[i].Span - 1
				update[i].Level[i].Forward = x.Level[i].Forward
			} else {
				update[i].Level[i].Span--
			}
		}

		if x.Level[0].Forward != nil {
			x.Level[0].Forward.Backward = x.Backward
		} else {
			sl.tail = x.Backward
		}

		for sl.level > 1 && sl.header.Level[sl.level-1].Forward == nil {
			sl.level--
		}
		sl.length--
		return true
	}
	return false
}

// insertNode 内部插入节点方法
func (sl *SkipList) insertNode(player *Player) {
	// 插入逻辑与Insert类似，但不加锁（因为外部已经加锁）
	// ... 省略具体实现
}
