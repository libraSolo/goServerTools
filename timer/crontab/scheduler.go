package crontab

import (
	"container/heap"
	"fmt"
	"os"
	"runtime/debug"
	"sync"
	"time"
)

// Scheduler 管理定时器。
type Scheduler struct {
	timers     timerHeap
	lock       sync.Mutex
	nextAddSeq uint
}

// NewScheduler 创建一个新的定时器调度器。
func NewScheduler() *Scheduler {
	s := &Scheduler{
		nextAddSeq: 1,
	}
	heap.Init(&s.timers)
	return s
}

// AddCallback 添加一个在指定持续时间后调用的回调。
func (s *Scheduler) AddCallback(d time.Duration, callback CallbackFunc) *Timer {
	return s.addTimer(d, callback, false)
}

// AddTimer 添加一个周期性调用回调的定时器。
func (s *Scheduler) AddTimer(d time.Duration, callback CallbackFunc) *Timer {
	if d < minTimerInterval {
		d = minTimerInterval
	}
	return s.addTimer(d, callback, true)
}

func (s *Scheduler) addTimer(d time.Duration, callback CallbackFunc, repeat bool) *Timer {
	t := &Timer{
		fireTime: time.Now().Add(d),
		interval: d,
		callback: callback,
		repeat:   repeat,
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	t.addSeq = s.nextAddSeq
	s.nextAddSeq++
	heap.Push(&s.timers, t)
	return t
}

// Tick 处理到期的定时器。
func (s *Scheduler) Tick() {
	now := time.Now()
	s.lock.Lock()
	defer s.lock.Unlock()

	for s.timers.Len() > 0 {
		t := s.timers[0]
		if t.fireTime.After(now) {
			break
		}

		heap.Pop(&s.timers)
		callback := t.callback
		if callback == nil {
			continue
		}

		if !t.repeat {
			t.callback = nil
		}

		// 解锁以运行回调，因为它可能会添加更多的定时器。
		s.lock.Unlock()
		runCallback(callback)
		s.lock.Lock()

		if t.repeat && t.IsActive() {
			t.fireTime = t.fireTime.Add(t.interval)
			if !t.fireTime.After(now) {
				t.fireTime = now.Add(t.interval)
			}
			t.addSeq = s.nextAddSeq
			s.nextAddSeq++
			heap.Push(&s.timers, t)
		}
	}
}

// Start 启动自计时程序。
func (s *Scheduler) Start(tickInterval time.Duration) {
	go s.selfTickRoutine(tickInterval)
}

func (s *Scheduler) selfTickRoutine(tickInterval time.Duration) {
	for {
		time.Sleep(tickInterval)
		s.Tick()
	}
}

func runCallback(callback CallbackFunc) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Fprintf(os.Stderr, "Callback panicked: %v\n", err)
			debug.PrintStack()
		}
	}()
	callback()
}