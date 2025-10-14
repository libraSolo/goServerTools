package post // 主线程投递模块：提供跨 Goroutine 将函数安全投递到“主游戏协程”执行的能力

import (
    "gwutils"
    "sync"
)

// PostCallback 表示可被投递到主线程执行的函数类型（无参数、无返回值）
type PostCallback func()

var (
    callbacks []PostCallback // 待执行的回调函数列表
    lock      sync.Mutex     // 互斥锁：保护 callbacks 的并发访问
)

// Post 将回调函数投递到主游戏协程，供稍后在主线程安全执行
//
// 该方法可能从其他 Goroutine 调用，因此使用互斥锁保证队列的并发安全
func Post(f PostCallback) {
    lock.Lock()
    callbacks = append(callbacks, f) // 追加到回调列表
    lock.Unlock()
}

// Tick 由主游戏协程调用：批量取出并执行所有已投递的回调函数
func Tick() {
    for { // 循环处理，直到当前批次没有新的回调待执行
        lock.Lock() // 加锁以安全读取并切换回调列表
        if len(callbacks) == 0 {
            lock.Unlock()
            break // 已无待执行回调，退出
        }
        // 在锁内将当前回调列表切换到副本，并清空原列表，以避免长时间持锁执行
        callbacksCopy := callbacks
        callbacks = make([]PostCallback, 0, len(callbacks)) // 清空队列，保留容量以减少分配
        lock.Unlock()

        for _, f := range callbacksCopy { // 逐个执行回调
            gwutils.RunPanicless(f) // 防护执行：回调若发生 panic 不影响后续回调
        }
    }
}
