package domain

import (
	"math/rand"
	"sync"
	"time"
)

// SkipListNode 跳表节点
type SkipListNode struct {
	Player   *Player         // 节点承载的玩家对象；头哨兵节点通常为 nil
	Backward *SkipListNode   // 第 0 层的后退指针，便于反向遍历与维护 tail
	Level    []SkipListLevel // 各层级的结构信息（前进指针与跨度），长度为该节点高度
}

// SkipListLevel 跳表层级
type SkipListLevel struct {
	Forward *SkipListNode // 本层的前进指针，指向同层的下一个节点；nil 表示该层末尾
	Span    int           // 从当前节点沿本层 Forward 跳到下一节点时跨过的“排名数量”，用于累计 rank
}

// SkipList 跳表
type SkipList struct {
	header *SkipListNode // 头哨兵节点（不承载玩家），拥有最大层级的 Level，所有查找/插入从此开始
	tail   *SkipListNode // 第 0 层的尾节点指针，便于末端操作与反向遍历
	length int           // 当前玩家节点数量，用于边界校验与复杂度估算
	level  int           // 跳表当前使用的最高层数（1..maxSkipListLevel），决定自顶向下查找的起始层
	mu     sync.RWMutex  // 并发读写锁：读操作使用 RLock，写操作（插入/删除/更新）使用 Lock，保障线程安全
}

const (
	maxSkipListLevel = 32
	skipListP        = 0.25
)

// NewSkipList 创建跳表
func NewSkipList() *SkipList {
	// 构造跳表：
	// - 初始最高层数为 1；
	// - header 为哨兵节点，预分配 maxSkipListLevel 层；
	// 复杂度：O(1)
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
	// 随机生成节点高度：以概率 p 递增层级，最大不超过 maxSkipListLevel。
	// 该策略使得期望复杂度保持在 O(log n)。
	level := 1
	for rand.Float32() < skipListP && level < maxSkipListLevel {
		level++
	}
	return level
}

// Delete 删除节点
func (sl *SkipList) Delete(playerID int64) bool {
	// 删除指定 ID 的节点：写锁保护。
	// 注意：避免访问 header.Player（为 nil）导致空指针；建议按 ID 精确匹配或复用 comparePlayers。
	// 复杂度：O(log n)
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
	// 返回给定排名区间 [start, end] 的玩家切片：读锁保护。
	// 先通过各层 span 快速定位到 start，再在第 0 层按节点顺序遍历到 end。
	// 复杂度：O(log n + k)，k 为区间长度。
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
    traversed := 0
    x := sl.header

    // 自顶向下定位到 rank=start 的前一位置，保持不越过 start
    for i := sl.level - 1; i >= 0; i-- {
        for x.Level[i].Forward != nil && (traversed+x.Level[i].Span) < start {
            traversed += x.Level[i].Span
            x = x.Level[i].Forward
        }
    }

    // 下一个节点即为 rank=start
    x = x.Level[0].Forward
    currentRank := traversed + 1

    // 遍历范围内的节点，直到 rank=end
    for x != nil && currentRank <= end {
        result = append(result, x.Player)
        x = x.Level[0].Forward
        currentRank++
    }

	return result
}

// Length 获取跳表长度
func (sl *SkipList) Length() int {
	// 返回跳表当前长度：读锁保护，O(1)
	sl.mu.RLock()
	defer sl.mu.RUnlock()
	return sl.length
}

// 比较函数 - 统一分数比较逻辑
func comparePlayers(p1, p2 *Player) int {
	// 排序规则：分数优先，其次更新时间（先更新者更前），最后 ID。
	// 返回值：1 表示 p1 更“高”（排在前面），-1 表示 p2 更高，0 表示完全相等。
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
	// 插入玩家节点：
	// - 自顶向下根据 comparePlayers 定位插入点；
	// - 随机生成高度并更新各层的 Forward 与 Span；
	// - 维护 Backward 与 tail。
	// 复杂度：期望 O(log n)
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
	// 获取指定玩家的排名：读锁保护。
	// 自顶向下按 span 累计 rank，最终在第 0 层确认是否命中该 ID。
	// 复杂度：O(log n)
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

// GetRankByPlayer 根据玩家分数键获取排名（按排序键查找）
// 使用与插入相同的比较逻辑 comparePlayers，自顶向下按 span 累计 rank。
// 复杂度：O(log n)
func (sl *SkipList) GetRankByPlayer(player *Player) (int, bool) {
    sl.mu.RLock()
    defer sl.mu.RUnlock()

    rank := 0
    x := sl.header

    for i := sl.level - 1; i >= 0; i-- {
        for x.Level[i].Forward != nil &&
            comparePlayers(x.Level[i].Forward.Player, player) > 0 {
            rank += x.Level[i].Span
            x = x.Level[i].Forward
        }
    }

    x = x.Level[0].Forward
    if x != nil && x.Player.ID == player.ID {
        return rank + 1, true
    }
    return 0, false
}

// UpdateScore 更新分数（需要删除再插入）
func (sl *SkipList) UpdateScore(player *Player, newScore int64) {
	// 更新分数：写锁保护。
	// 流程：删除旧节点 -> 更新分数与时间 -> 无锁内部插入（外部已加锁）。
	// 保证有序性与排名正确。
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
	// 内部删除：按 ID 精确定位并维护各层 span 与 Forward。
	// 若删除的是尾节点，更新 tail；必要时降低最高层 level。
	// 复杂度：O(log n)
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
	// 内部插入：与 Insert 类似，但不加锁（调用方已加锁）。
	// 维护各层 Forward/Span 与 Backward/tail。
	// 复杂度：期望 O(log n)

	update := make([]*SkipListNode, maxSkipListLevel)
	rank := make([]int, maxSkipListLevel)
	x := sl.header

	// 自顶向下定位插入点
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

	// 随机生成层级并可能提升最高层
	level := sl.randomLevel()
	if level > sl.level {
		for i := sl.level; i < level; i++ {
			rank[i] = 0
			update[i] = sl.header
			update[i].Level[i].Span = sl.length
		}
		sl.level = level
	}

	// 创建新节点
	x = &SkipListNode{
		Player: player,
		Level:  make([]SkipListLevel, level),
	}

	// 更新各层指针与 span
	for i := 0; i < level; i++ {
		x.Level[i].Forward = update[i].Level[i].Forward
		update[i].Level[i].Forward = x
		x.Level[i].Span = update[i].Level[i].Span - (rank[0] - rank[i])
		update[i].Level[i].Span = (rank[0] - rank[i]) + 1
	}

	// 更新更高层的 span（新节点未触达的层）
	for i := level; i < sl.level; i++ {
		update[i].Level[i].Span++
	}

	// 维护后退指针与尾指针
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
