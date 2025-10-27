package crontab

import (
	"fmt"
	"gwutils"
	"time"
)

const (
	timeOffset = time.Second * 5
)

// Handle 是定时任务的句柄类型，用于取消注册的任务
type Handle int

var (
	cronScheduler    = NewScheduler()
	cancelledHandles = []Handle{}          // 待取消的任务句柄列表
	entries          = map[Handle]*entry{} // 所有注册的定时任务条目
	nextHandle       = Handle(1)           // 下一个可用的任务句柄
)

// entry 表示一个定时任务条目
type entry struct {
	minute, hour, day, month, dayofweek int    // 定时任务的时间参数
	callback                            func() // 定时任务的回调函数
}

func (entry *entry) match(minute int, hour int, day int, month time.Month, weekday time.Weekday) bool {
	// matchField 是一个辅助函数，用于检查给定的时间值是否与 cron 条目的字段匹配。
	// 负数的 entryValue 表示步长值（例如，*/5 存储为 -5）。
	matchField := func(entryValue, timeValue int) bool {
		if entryValue < 0 {
			return timeValue%-entryValue == 0
		}
		return entryValue == timeValue
	}

	if !matchField(entry.minute, minute) ||
		!matchField(entry.hour, hour) ||
		!matchField(entry.day, day) ||
		!matchField(entry.month, int(month)) {
		return false
	}

	if entry.dayofweek >= 0 { // 值为 -1 表示一周中的任意一天。
		// 在 cron 规范中，0 和 7 都可以表示星期日。
		// 我们将 7 规范化为 0，以与 Go 的 time.Sunday 对齐。
		cronDay := entry.dayofweek
		if cronDay == 7 {
			cronDay = 0
		}
		if cronDay != int(weekday) {
			return false
		}
	}

	return true
}

// Register 注册一个定时任务，当时间条件满足时执行回调函数
// 参数说明：
//
//	minute: 分钟（0-59），负数表示每隔-minute分钟执行一次
//	hour: 小时（0-23），负数表示每隔-hour小时执行一次
//	day: 日（1-31），负数表示每隔-day天执行一次
//	month: 月（1-12），负数表示每隔-month月执行一次
//	dayofweek: 星期几（0-7，0和7均为周日），负数表示每隔-dayofweek周执行一次
//	cb: 时间条件满足时执行的回调函数
//
// 返回值：任务句柄，可用于取消任务
func Register(minute, hour, day, month, dayofweek int, cb func()) Handle {
	validateTime(minute, hour, day, month, dayofweek)

	h := genNextHandle()
	entries[h] = &entry{
		minute:    minute,
		hour:      hour,
		day:       day,
		month:     month,
		dayofweek: dayofweek,
		callback:  cb,
	}
	return h
}

func validateTime(minute, hour, day, month, dayofweek int) bool {
	if minute > 59 || minute < -60 {
		return false
	}

	if hour > 23 || hour < -24 {
		return false
	}
	if day > 31 || day < -31 || day == 0 {
		return false
	}
	if month > 12 || month < -12 || month == 0 {
		return false
	}
	if dayofweek > 7 || dayofweek < -1 {
		return false
	}

	return true
}

func genNextHandle() (h Handle) {
	h, nextHandle = nextHandle, nextHandle+1
	return
}

// Unregister 取消注册一个定时任务
func (h Handle) Unregister() {
	cancelledHandles = append(cancelledHandles, h)
}

// reset resets the crontab state for testing.
func reset() {
	entries = make(map[Handle]*entry)
	nextHandle = Handle(1)
	cancelledHandles = []Handle{}
}

func unregisterCancelledHandles() {
	for _, h := range cancelledHandles {
		fmt.Printf("unregisterCancelledHandles: cancelling %d", h)
		delete(entries, h)
	}
	cancelledHandles = nil
}

// Initialize 初始化定时任务模块，由引擎调用
func Initialize() {
	now := time.Now()
	sec := now.Second()
	var d time.Duration
	if time.Second*time.Duration(sec) < timeOffset {
		d = timeOffset - time.Second*time.Duration(sec)
	} else {
		d = time.Second*time.Duration(60-sec) + timeOffset
	}

	d -= time.Nanosecond * time.Duration(now.Nanosecond())
	fmt.Printf("current time is %s, will setup repeat time after %s", now, d)
	cronScheduler.AddCallback(d, func() {
		setupRepeatTimer()
		checkNow()
	})
	cronScheduler.Start(time.Second) // Start the scheduler to tick every second
}

func check(now time.Time) {
	unregisterCancelledHandles()

	fmt.Printf("Crontab: checking %d callbacks ...", len(entries))
	dayofweek, month, day, hour, minute := now.Weekday(), now.Month(), now.Day(), now.Hour(), now.Minute()

	for _, entry := range entries {
		if entry.match(minute, hour, day, month, dayofweek) {
			gwutils.RunPanicless(entry.callback)
		}
	}

	unregisterCancelledHandles()
}

func checkNow() {
	check(time.Now())
}

func setupRepeatTimer() {
	fmt.Printf("Crontab: setup repeat timer at time %s", time.Now())
	cronScheduler.AddTimer(time.Minute, checkNow)
}
