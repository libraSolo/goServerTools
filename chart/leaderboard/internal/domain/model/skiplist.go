package model

import (
	"math/rand"
	"time"
)

const (
	maxLevel = 32
	 p        = 0.25
)

type ( 
	// Node 表示跳表中的一个节点。
	Node struct {
		Player   *Player
		backward *Node
		level    []*Level
	}

	// Level 表示节点在特定层的指针。
	Level struct {
		forward *Node
		span    int64
	}

	// SkipList 是一个跳表数据结构。
	SkipList struct {
		header *Node
		tail   *Node
		length int64
		level  int
	}
)

// NewSkipList 创建一个新的跳表。
func NewSkipList() *SkipList {
	rand.Seed(time.Now().UnixNano())
	return &SkipList{
		header: &Node{
			Player: &Player{Score: -1}, // Header node with a sentinel score
			level:  make([]*Level, maxLevel),
		},
		level: 1,
	}
}

// randomLevel 生成一个随机的层数。
func randomLevel() int {
	level := 1
	for rand.Float64() < p && level < maxLevel {
		level++
	}
	return level
}

// Insert 插入一个新玩家到跳表中。
func (sl *SkipList) Insert(player *Player) *Node {
	update := make([]*Node, maxLevel)
	rank := make([]int64, maxLevel)
	x := sl.header

	for i := sl.level - 1; i >= 0; i-- {
		// store rank that is crossed to reach the insert position
		if i == sl.level-1 {
			rank[i] = 0
		} else {
			rank[i] = rank[i+1]
		}
		for x.level[i] != nil && (x.level[i].forward.Player.Score > player.Score || (x.level[i].forward.Player.Score == player.Score && x.level[i].forward.Player.ID < player.ID)) {
			rank[i] += x.level[i].span
			x = x.level[i].forward
		}
		update[i] = x
	}

	level := randomLevel()
	if level > sl.level {
		for i := sl.level; i < level; i++ {
			rank[i] = 0
			update[i] = sl.header
			update[i].level[i] = &Level{span: sl.length}
		}
		sl.level = level
	}

	x = &Node{Player: player, level: make([]*Level, level)}
	for i := 0; i < level; i++ {
		x.level[i] = &Level{}
		if update[i].level[i] == nil {
			update[i].level[i] = &Level{}
		}
		x.level[i].forward = update[i].level[i].forward
		update[i].level[i].forward = x

		// update span covered by update[i] as x is inserted here
		x.level[i].span = update[i].level[i].span - (rank[0] - rank[i])
		update[i].level[i].span = (rank[0] - rank[i]) + 1
	}

	// increment span for untouched levels
	for i := level; i < sl.level; i++ {
		if update[i].level[i] != nil {
			update[i].level[i].span++
		}
	}

	if update[0] == sl.header {
		x.backward = nil
	} else {
		x.backward = update[0]
	}
	if x.level[0].forward != nil {
		x.level[0].forward.backward = x
	} else {
		sl.tail = x
	}
	sl.length++
	return x
}

// GetRank 获取玩家的排名。
func (sl *SkipList) GetRank(score int64, id int64) int64 {
	var rank int64 = 0
	x := sl.header
	for i := sl.level - 1; i >= 0; i-- {
		for x.level[i] != nil && (x.level[i].forward.Player.Score > score || (x.level[i].forward.Player.Score == score && x.level[i].forward.Player.ID < id)) {
			rank += x.level[i].span
			x = x.level[i].forward
		}
	}
	return rank + 1
}

// GetElementByRank 通过排名获取玩家。
func (sl *SkipList) GetElementByRank(rank int64) *Node {
	if rank < 1 || rank > sl.length {
		return nil
	}
	var traversed int64 = 0
	x := sl.header
	for i := sl.level - 1; i >= 0; i-- {
		for x.level[i] != nil && (traversed+x.level[i].span) <= rank {
			traversed += x.level[i].span
			x = x.level[i].forward
		}
		if traversed == rank {
			return x
		}
	}
	return nil
}

// Delete 从跳表中删除一个节点。
func (sl *SkipList) Delete(score int64, id int64) {
	update := make([]*Node, maxLevel)
	x := sl.header

	for i := sl.level - 1; i >= 0; i-- {
		for x.level[i] != nil && (x.level[i].forward.Player.Score > score || (x.level[i].forward.Player.Score == score && x.level[i].forward.Player.ID < id)) {
			x = x.level[i].forward
		}
		update[i] = x
	}

	x = x.level[0].forward

	if x != nil && x.Player.Score == score && x.Player.ID == id {
		for i := 0; i < sl.level; i++ {
			if update[i].level[i] != nil && update[i].level[i].forward == x {
				if update[i].level[i] != nil && update[i].level[i].forward == x {
				update[i].level[i].span += x.level[i].span - 1
				update[i].level[i].forward = x.level[i].forward
			} else if update[i].level[i] != nil {
				update[i].level[i].span--
			}
		}
		if x.level[0].forward != nil {
			x.level[0].forward.backward = x.backward
		} else {
			sl.tail = x.backward
		}
		for sl.level > 1 && sl.header.level[sl.level-1] == nil {
			sl.level--
		}
		sl.length--
	}
}
}