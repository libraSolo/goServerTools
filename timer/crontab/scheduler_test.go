package crontab

import (
	"fmt"
	"testing"
	"time"
)

// TestScheduler_Start 演示了如何使用 Start() 方法来启动一个自动运行的调度器。
func TestScheduler_Start(t *testing.T) {
	scheduler := NewScheduler()

	// 添加一个一次性回调
	scheduler.AddCallback(100*time.Millisecond, func() {
		fmt.Println("Callback executed")
	})

	// 添加一个周期性定时器
	timer := scheduler.AddTimer(200*time.Millisecond, func() {
		fmt.Println("Timer ticked")
	})

	// 启动调度器，tick 间隔为 50 毫秒
	scheduler.Start(50 * time.Millisecond)

	// 等待足够长的时间以观察输出
	time.Sleep(1000 * time.Millisecond)

	// 取消周期性定时器
	timer.Cancel()

	// 再等待一段时间，确保定时器已停止
	time.Sleep(1500 * time.Millisecond)

	fmt.Println("TestScheduler_Start finished")
}

// TestScheduler_ManualTick 演示了如何通过手动调用 Tick() 方法来处理定时器。
func TestScheduler_ManualTick(t *testing.T) {
	scheduler := NewScheduler()

	// 添加一个一次性回调
	scheduler.AddCallback(100*time.Millisecond, func() {
		fmt.Println("Manual tick: Callback executed")
	})

	// 模拟时间的流逝和手动的 tick
	for i := 0; i < 5; i++ {
		time.Sleep(50 * time.Millisecond)
		scheduler.Tick()
		fmt.Println("Manual tick")
	}

	fmt.Println("TestScheduler_ManualTick finished")
}
