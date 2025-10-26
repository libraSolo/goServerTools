// DelayQueue 模块：基于最小堆的延时队列，实现“到期后投递”逻辑。
// 用法概要：启动 Poll 循环后，调用 Offer 写入元素与到期时间，元素到期时会从 C 通道输出。
package timeWheel

import (
	"container/heap"
	"sync"
	"sync/atomic"
	"time"
)

// DelayQueue 小根堆实现的优先级队列
// DelayQueue 小根堆实现的优先级队列：
// - C：元素到期后被发送到此通道
// - Offer：写入元素及到期时间
// - Poll：循环等待到期元素并投递到 C
// 并发安全：内部使用互斥锁与原子标志协调“睡眠/唤醒”。
type DelayQueue struct {
	C chan interface{}

	mu sync.Mutex
	pq priorityQueue

	sleeping int32
	wakeupC  chan struct{}
}

// NewDelayQueue 创建一个初始容量为 size 的延时队列。
// 注意：需要在独立 goroutine 中调用 Poll 才会持续投递到期元素。
func NewDelayQueue(size int) *DelayQueue {
	return &DelayQueue{
		C:       make(chan interface{}),
		pq:      newPriorityQueue(size),
		wakeupC: make(chan struct{}),
	}
}

// Offer 写入一个指定到期时间的元素到当前的延时队列。
// 参数：elem 为任意对象（典型为 *Bucket），expiration 为毫秒时间戳。
// 行为：若新元素成为堆顶且 Poll 线程处于“睡眠”，则通过 wakeupC 唤醒它。
func (dq *DelayQueue) Offer(elem interface{}, expiration int64) {
	item := &item{
		Value:    elem,
		Priority: expiration,
	}

	dq.mu.Lock()
	heap.Push(&dq.pq, item)
	index := item.Index
	dq.mu.Unlock()

	if index == 0 {
		if atomic.CompareAndSwapInt32(&dq.sleeping, 1, 0) {
			dq.wakeupC <- struct{}{}
		}
	}
}

// Poll 无限循环的获取一个元素，并按到期时间阻塞/唤醒。
// 参数：
// - exitC：退出信号，关闭或收到值后退出 Poll 循环
// - nowF：提供“当前毫秒时间”的函数，便于注入与测试
// 行为：
// - 若队首未到期，可能休眠 delta 毫秒，或被新堆顶元素唤醒
// - 到期元素会被发送到 C 通道供上层消费

func (dq *DelayQueue) Poll(exitC chan struct{}, nowF func() int64) {
	defer atomic.StoreInt32(&dq.sleeping, 0)
	for {
		now := nowF()

		dq.mu.Lock()
		item, delta := dq.pq.PeekAndShift(now)
		if item == nil {
			atomic.StoreInt32(&dq.sleeping, 1)
		}
		dq.mu.Unlock()

		if item == nil {
			if delta == 0 {
				select {
				case <-dq.wakeupC:
					continue
				case <-exitC:
					return
				}
			} else if delta > 0 {
				select {
				case <-dq.wakeupC:
					continue
				case <-time.After(time.Duration(delta) * time.Millisecond):
					if atomic.SwapInt32(&dq.sleeping, 0) == 0 {
						<-dq.wakeupC
					}
					continue
				case <-exitC:
					return
				}
			}
		}

		select {
		case dq.C <- item.Value:
		case <-exitC:
			return
		}
	}
}
