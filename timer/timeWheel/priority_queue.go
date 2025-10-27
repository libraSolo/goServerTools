// Package timeWheel 的优先队列子模块
// 本文件实现一个基于最小堆的优先队列，用于延时队列按到期时间排序元素。
// 使用场景：与 DelayQueue 搭配，实现毫秒级到期的元素管理。
package timeWheel

import (
	"container/heap"
)

// item 表示队列中的元素，Priority 通常为到期时间（毫秒）。
// Value 保存具体对象（例如 Bucket），Index 为在堆中的位置。
type item struct {
	Value    interface{}
	Priority int64
	Index    int
}

// priorityQueue 为最小堆容器，满足 heap.Interface。
// 仅存储 *item，以 Priority 作为比较依据。
type priorityQueue []*item

// newPriorityQueue 创建一个具有初始 capacity 的优先队列。
// 注意：内部会在 Push/Pop 时按需扩缩容，避免频繁分配。
func newPriorityQueue(capacity int) priorityQueue {
	return make(priorityQueue, 0, capacity)
}

func (pq priorityQueue) Len() int {
	return len(pq)
}

func (pq priorityQueue) Less(i, j int) bool {
	return pq[i].Priority < pq[j].Priority
}

func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].Index = i
	pq[j].Index = j
}

func (pq *priorityQueue) Push(x interface{}) {
	n := len(*pq)
	c := cap(*pq)
	if n+1 > c {
		newCap := max(c * 2, 8)
		npq := make(priorityQueue, n, newCap)
		copy(npq, *pq)
		*pq = npq
	}
	*pq = (*pq)[0 : n+1]
	item := x.(*item)
	item.Index = n
	(*pq)[n] = item
}

func (pq *priorityQueue) Pop() interface{} {
	n := len(*pq)
	c := cap(*pq)
	if n < (c/2) && c > 25 {
		newCap := c / 2
		if newCap < n {
			newCap = n
		}

		npq := make(priorityQueue, n, newCap)
		copy(npq, *pq)
		*pq = npq
	}
	item := (*pq)[n-1]
	item.Index = -1
	*pq = (*pq)[0 : n-1]
	return item
}

// PeekAndShift 返回当前队首（最小 Priority）的元素：
// - 若队首元素的 Priority <= max，则移除并返回该元素；第二个返回值为 0。
// - 若队首元素尚未到期，则不移除，返回 nil 与剩余等待时间（Priority - max）。
func (pq *priorityQueue) PeekAndShift(max int64) (*item, int64) {
	if pq.Len() == 0 {
		return nil, 0
	}

	item := (*pq)[0]
	if item.Priority > max {
		return nil, item.Priority - max
	}
	heap.Remove(pq, 0)

	return item, 0
}