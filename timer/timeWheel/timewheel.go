// 时间轮（TimeWheel）模块：层级时间轮 + 延时队列，实现高并发定时任务调度。
// 典型用法：
//
//	dq := NewDelayQueue(64)
//	tw := NewTimeWheel(100, 512, time.Now().UnixNano()/1e6, dq) // tick=100ms, 512个格
//	tw.Start()
//	tw.tryAdd(&TimerTaskEntity{DelayTime: time.Now().UnixNano()/1e6 + 500, Task: func(){ /* do work */ }})
//	// ... 业务逻辑 ...
//	tw.Stop()
//
// 说明：
// - tryAdd：若任务在当前 tick 内到期，直接执行；否则加入对应 Bucket 或溢出到上层轮。
// - Start：启动两个后台循环：一个维护 DelayQueue 的到期投递，一个处理桶到期后的降级与执行。
// - Stop：关闭并等待后台循环退出，保证资源回收。
package timeWheel

import (
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

// TimeWheel 时间轮：
// - tick：每个时间格的跨度（毫秒）
// - wheelSize：时间轮包含的格子数，总跨度为 tick*wheelSize
// - queue：共享的 DelayQueue（所有层级时间轮使用同一个队列）
// - overflow：更高层的时间轮，用于承载更长延时的任务
// - currentTime：当前对齐到 tick 的时间
// TimeWheel 时间轮
type TimeWheel struct {
	tick        int64       // 基本时间跨度
	wheelSize   int64       // 时间轮大小
	interval    int64       // 时间轮总跨度
	buckets     []*Bucket   // 时间格
	queue       *DelayQueue // 延时队列
	overflow    *TimeWheel  // 上层时间轮
	currentTime int64       // 当前时间
	exitC       chan struct{}
	waitGroup   sync.WaitGroup
}

// NewTimeWheel 创建一个时间轮。
// 参数：
// - tick：时间格跨度（毫秒），例如 100 表示 100ms
// - wheelSize：格子数量，例如 512
// - startMs：起始时间（毫秒），会被按 tick 对齐
// - queue：共享的延时队列实例，用于所有层级轮的到期调度
func NewTimeWheel(tick int64, wheelSize int64, startMs int64, queue *DelayQueue) *TimeWheel {
	buckets := make([]*Bucket, wheelSize)
	for i := range buckets {
		buckets[i] = newBucket()
	}
	return &TimeWheel{
		tick:        tick,
		wheelSize:   wheelSize,
		interval:    tick * wheelSize,
		buckets:     buckets,
		queue:       queue,
		currentTime: truncate(startMs, tick),
		exitC:       make(chan struct{}),
	}
}

// add 尝试将任务加入当前时间轮：
// - 若任务到期在 [currentTime+tick, currentTime+interval) 范围内，落入当前轮的某个 Bucket
// - 否则溢出到上层时间轮（必要时按需创建 overflow）
// 返回：是否成功加入到某个时间轮（未到期时返回 true；否则 false 表示应直接执行）
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

// tryAdd 将任务尝试加入时间轮；若已到执行窗口内，则直接异步执行。
func (tw *TimeWheel) tryAdd(t *TimerTaskEntity) {
	if !tw.add(t) {
		go t.Task()
	}
}

// advanceClock 推进时间轮的当前时间到给定 timeMs 所在的对齐刻度，并联动上层轮。
func (tw *TimeWheel) advanceClock(timeMs int64) {
	if timeMs >= tw.currentTime+tw.tick {
		tw.currentTime = truncate(timeMs, tw.tick)
		if tw.overflow != nil {
			tw.overflow.advanceClock(tw.currentTime)
		}
	}
}

// Start 启动时间轮：
// - 后台 goroutine1：维护 DelayQueue 的 Poll 循环
// - 后台 goroutine2：处理到期桶（Bucket）的降级与任务分发/执行
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
				b.Flush(tw.tryAdd)
			}
		}
	}()
}

// Stop 停止时间轮：
// 关闭退出通道并等待后台 goroutine 退出，确保有序回收资源。
func (tw *TimeWheel) Stop() {
	close(tw.exitC)
	tw.waitGroup.Wait()
}

// truncate 将时间 x 按步长 m 对齐到下一个不超过 x 的整刻度。
// 用于确保 currentTime 与 bucket 过期时间严格按 tick 对齐，避免抖动。
func truncate(x, m int64) int64 {
	if m <= 0 {
		return x
	}
	return x - x%m
}
