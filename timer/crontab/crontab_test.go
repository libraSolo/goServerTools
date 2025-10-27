package crontab

import (
	"sync"
	"testing"
	"time"
)

func TestCrontab_DirectCheck(t *testing.T) {
	reset()

	var lock sync.Mutex
	triggered := false

	// 注册一个每分钟都会触发的任务
	h := Register(-1, -1, -1, -1, -1, func() {
		lock.Lock()
		triggered = true
		lock.Unlock()
	})

	// 直接调用 check 函数
	check(time.Now())

	lock.Lock()
	if !triggered {
		t.Fatal("Callback was not triggered")
	}
	// 重置标志
	triggered = false
	lock.Unlock()

	// 取消注册任务
	h.Unregister()

	// 再次直接调用 check 函数
	check(time.Now())

	lock.Lock()
	if triggered {
		t.Fatal("Callback was triggered after unregistering")
	}
	lock.Unlock()
}

func TestCrontab_SpecificTime(t *testing.T) {
	reset()

	var lock sync.Mutex
	triggered1 := false
	triggered2 := false

	// 注册一个在 10:30 触发的任务
	Register(30, 10, -1, -1, -1, func() {
		lock.Lock()
		triggered1 = true
		lock.Unlock()
	})

	// 注册一个在 11:30 触发的任务
	Register(30, 11, -1, -1, -1, func() {
		lock.Lock()
		triggered2 = true
		lock.Unlock()
	})

	// 模拟时间 2025-10-27 10:30:00
	now := time.Date(2025, 10, 27, 10, 30, 0, 0, time.UTC)
	check(now)

	lock.Lock()
	if !triggered1 {
		t.Fatal("Callback 1 was not triggered at the specific time")
	}
	if triggered2 {
		t.Fatal("Callback 2 was triggered at the wrong time")
	}
	lock.Unlock()
}