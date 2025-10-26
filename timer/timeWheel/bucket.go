package timeWheel

import (
	"container/list"
	"sync"
	"sync/atomic"
)

// Bucket 时间格：
// 存放处于同一到期时间片的延时任务列表，过期后由时间轮批量“降级/重插”到低层或执行。
// 线程安全：内部使用互斥锁保护任务链表与任务关联的更新。
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

// Expiration 返回当前桶的过期时间（毫秒），若为 -1 表示未设置/无效。
func (b *Bucket) Expiration() int64 {
	return atomic.LoadInt64(&b.expiration)
}

// SetExpiration 设置桶的过期时间；返回值表示过期时间是否发生变化。
// 若设置为新的过期时间，时间轮会将该桶加入 DelayQueue。
func (b *Bucket) SetExpiration(expiration int64) bool {
	return atomic.SwapInt64(&b.expiration, expiration) != expiration
}

// Add 将任务加入桶的链表，并关联任务所在桶与双向链表节点。
func (b *Bucket) Add(t *TimerTaskEntity) {
	b.mu.Lock()
	defer b.mu.Unlock()

	element := b.tasks.PushBack(t)
	t.setBucket(b)
	t.element = element
}

// Remove 尝试从桶中移除指定任务，返回是否移除成功。
// 注意：仅当任务当前确实属于该桶时才会移除并清理关联。
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

// Flush 任务降级：
// 当桶到期时，将桶内所有任务集中取出并解除与桶的关联，随后在解锁后逐个回调 reinsert。
// 这样可以避免对 Remove 的重入导致死锁，并减少长时间持锁对其他操作的影响。
// 任务降级
func (b *Bucket) Flush(reinsert func(t *TimerTaskEntity)) {
	// 收集并移除当前桶中的任务，避免在持锁情况下再次调用 Remove 造成重入死锁
	b.mu.Lock()
	var toReinsert []*TimerTaskEntity
	for e := b.tasks.Front(); e != nil; {
		next := e.Next()
		t := e.Value.(*TimerTaskEntity)
		// 在持锁状态下直接移除，清理任务与桶的关联
		b.tasks.Remove(e)
		t.setBucket(nil)
		t.element = nil
		toReinsert = append(toReinsert, t)
		e = next
	}
	// 当前桶所有延时任务降级完成后，该桶过期时间重置为-1，该桶不再有效
	b.SetExpiration(-1)
	b.mu.Unlock()

	// 在释放锁后执行降级/重插，避免长时间持锁影响其他桶操作
	for _, t := range toReinsert {
		reinsert(t)
	}
}