package timeWheel

import (
	"container/list"
	"sync/atomic"
	"unsafe"
)

// TimerTaskEntity 延时任务实体：
// - DelayTime：任务目标执行时间点（毫秒时间戳）
// - Task：实际执行的函数
// 使用：通常由 TimeWheel 尝试加入；若 DelayTime 已进入当前 tick，则直接执行 Task。
// 并发：内部通过原子指针记录所在 Bucket，并在 Stop/Remove 时安全更新。
//
// 示例：
// t := &TimerTaskEntity{DelayTime: now+500, Task: func(){ fmt.Println("run") }}
// tw.tryAdd(t) // 若等待时间在当前时间轮范围内，则进入对应 Bucket；否则溢出到上层轮。
// t.Stop()     // 若尚未执行，尝试从所在 Bucket 移除。
//
// 注：Stop 仅保证“尝试取消”，若任务已被提升/执行，Stop 返回可能为 false。
// TimerTaskEntity 延时任务
type TimerTaskEntity struct {
	DelayTime int64 // 延时时间
	Task      func()

	b unsafe.Pointer     // type: *bucket  保存当前延时任务所在的时间格，使用桶指针，可通过原子操作并发更新/读取
	
	element *list.Element // 延时任务所在的双向链表中的节点元素
}

// getBucket 获取任务当前所在的时间格（Bucket），可能为 nil。
func (t *TimerTaskEntity) getBucket() *Bucket {
	return (*Bucket)(atomic.LoadPointer(&t.b))
}

// setBucket 更新任务所在时间格（Bucket），使用原子指针保证并发安全。
func (t *TimerTaskEntity) setBucket(b *Bucket) {
	atomic.StorePointer(&t.b, unsafe.Pointer(b))
}

// Stop 停止延时任务的执行：
// - 若任务仍在某个 Bucket 中，尝试调用 Bucket.Remove 将其移除
// - 若任务已被提升到低层时间轮或正在执行，可能无法停止
// 返回：是否成功从 Bucket 中移除（表示取消成功）
// Stop 停止延时任务的执行
func (t *TimerTaskEntity) Stop() bool {
	stopped := false
	for b := t.getBucket(); b != nil; b = t.getBucket() {
		// 如果时间格尚未过期/执行，则从时间格中删除这个延时任务
		stopped = b.Remove(t)
	}
	return stopped
}