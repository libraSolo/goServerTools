package crontab

import (
	"time"
)

const (
	// minTimerInterval 定义了定时器的最小间隔。
	minTimerInterval = 1 * time.Millisecond
)

// CallbackFunc 定义了回调函数的类型。
type CallbackFunc func()

// Timer 表示一个已调度的任务。
type Timer struct {
	fireTime time.Time
	interval time.Duration
	callback CallbackFunc
	repeat   bool
	addSeq   uint
}

// Cancel 取消定时器。
func (t *Timer) Cancel() {
	t.callback = nil
}

// IsActive 检查定时器是否仍处于活动状态。
func (t *Timer) IsActive() bool {
	return t.callback != nil
}
