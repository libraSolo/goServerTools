package timeWheel

import (
	"container/heap"
	"container/list"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

type item struct {
	Value    interface{}
	Priority int64
	Index    int
}

type priorityQueue []*item

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
		npq := make(priorityQueue, n, c*2)
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
		npq := make(priorityQueue, n, c/2)
		copy(npq, *pq)
		*pq = npq
	}
	item := (*pq)[n-1]
	item.Index = -1
	*pq = (*pq)[0 : n-1]
	return item
}

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

// DelayQueue 小根堆实现的优先级队列
type DelayQueue struct {
	C chan interface{}

	mu sync.Mutex
	pq priorityQueue

	sleeping int32
	wakeupC  chan struct{}
}

func NewDelayQueue(size int) *DelayQueue {
	return &DelayQueue{
		C:       make(chan interface{}),
		pq:      newPriorityQueue(size),
		wakeupC: make(chan struct{}),
	}
}

// Offer 写入一个指定到期时间的元素到当前的延时队列
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

// Poll 无限循环的获取一个元素
func (dq *DelayQueue) Poll(exitC chan struct{}, nowF func() int64) {
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
					goto exit
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
					goto exit
				}
			}
		}

		select {
		case dq.C <- item.Value:
		case <-exitC:
			goto exit
		}
	}

exit:
	atomic.StoreInt32(&dq.sleeping, 0)
}

// TimerTaskEntity 延时任务
type TimerTaskEntity struct {
	DelayTime int64 // 延时时间
	Task      func()

	b unsafe.Pointer // type: *bucket  保存当前延时任务所在的时间格，使用桶指针，可通过原子操作并发更新/读取

	element *list.Element // 延时任务所在的双向链表中的节点元素

}

func (t *TimerTaskEntity) getBucket() *Bucket {
	return (*Bucket)(atomic.LoadPointer(&t.b))
}

func (t *TimerTaskEntity) setBucket(b *Bucket) {
	atomic.StorePointer(&t.b, unsafe.Pointer(b))
}

// Stop 停止延时任务的执行
func (t *TimerTaskEntity) Stop() bool {
	stopped := false
	for b := t.getBucket(); b != nil; b = t.getBucket() {
		// 如果时间格尚未过期/执行，则从时间格中删除这个延时任务
		stopped = b.Remove(t)
	}
	return stopped
}

// Bucket 时间格
type Bucket struct {
	// 过期时间
	expiration int64

	// 任务列表
	tasks *list.List

	// 互斥锁
	mu sync.Mutex
}

func newBucket() *Bucket {
	return &Bucket{
		expiration: -1,
		tasks:      list.New(),
	}
}

func (b *Bucket) Expiration() int64 {
	return atomic.LoadInt64(&b.expiration)
}

func (b *Bucket) SetExpiration(expiration int64) bool {
	return atomic.SwapInt64(&b.expiration, expiration) != expiration
}

func (b *Bucket) Add(t *TimerTaskEntity) {
	b.mu.Lock()
	defer b.mu.Unlock()

	element := b.tasks.PushBack(t)
	t.setBucket(b)
	t.element = element
}

func (b *Bucket) Remove(t *TimerTaskEntity) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	if t.getBucket() != b {
		return false
	}
	b.tasks.Remove(t.element)
	t.setBucket(nil)
	t.element = nil
	return true
}

func (b *Bucket) Flush(reinsert func(t *TimerTaskEntity)) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for e := b.tasks.Front(); e != nil; {
		next := e.Next()
		t := e.Value.(*TimerTaskEntity)
		b.Remove(t)
		reinsert(t)
		e = next
	}
	b.SetExpiration(-1)
}

// TimeWheel 时间轮
type TimeWheel struct {
	tick      int64        // 基本时间跨度
	wheelSize int64        // 时间轮大小
	interval  int64        // 时间轮总跨度
	buckets   []*Bucket    // 时间格
	queue     *DelayQueue  // 延时队列
	overflow  *TimeWheel   // 上层时间轮
	currentTime int64      // 当前时间
	exitC     chan struct{}
	waitGroup sync.WaitGroup
}

func NewTimeWheel(tick int64, wheelSize int64, startMs int64, queue *DelayQueue) *TimeWheel {
	buckets := make([]*Bucket, wheelSize)
	for i := range buckets {
		buckets[i] = newBucket()
	}
	return &TimeWheel{
		tick:      tick,
		wheelSize: wheelSize,
		interval:  tick * wheelSize,
		buckets:   buckets,
		queue:     queue,
		currentTime: truncate(startMs, tick),
		exitC:     make(chan struct{}),
	}
}

func (tw *TimeWheel) add(t *TimerTaskEntity) bool {
	currentTime := atomic.LoadInt64(&tw.currentTime)
	if t.DelayTime < currentTime+tw.tick {
		return false
	} else if t.DelayTime < currentTime+tw.interval {
		virtualID := t.DelayTime / tw.tick
		bucket := tw.buckets[virtualID%tw.wheelSize]
		bucket.Add(t)
		if bucket.SetExpiration(virtualID * tw.tick) {
			tw.queue.Offer(bucket, bucket.Expiration())
		}
		return true
	} else {
		if tw.overflow == nil {
			atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&tw.overflow)), nil, unsafe.Pointer(NewTimeWheel(tw.interval, tw.wheelSize, currentTime, tw.queue)))
		}
		return tw.overflow.add(t)
	}
}

func (tw *TimeWheel) advanceClock(timeMs int64) {
	if timeMs >= tw.currentTime+tw.tick {
		tw.currentTime = truncate(timeMs, tw.tick)
		if tw.overflow != nil {
			tw.overflow.advanceClock(tw.currentTime)
		}
	}
}

func (tw *TimeWheel) Start() {
	tw.waitGroup.Add(1)
	go func() {
		defer tw.waitGroup.Done()
		tw.queue.Poll(tw.exitC, func() int64 {
			return time.Now().UnixNano() / 1e6
		})
	}()

	tw.waitGroup.Add(1)
	go func() {
		defer tw.waitGroup.Done()
		for {
			select {
			case <-tw.exitC:
				return
			case elem := <-tw.queue.C:
				b := elem.(*Bucket)
				tw.advanceClock(b.Expiration())
				b.Flush(tw.add)
			}
		}
	}()
}

func (tw *TimeWheel) Stop() {
	close(tw.exitC)
	tw.waitGroup.Wait()
}

func truncate(x, m int64) int64 {
	if m <= 0 {
		return x
	}
	return x - x%m
}