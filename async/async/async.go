package async // 异步任务模块：提供在独立 Goroutine 中执行任务，并将结果回投到游戏主线程

import (
    "fmt"
    "gwutils"
    "post"
    "sync"
)

var (
    numAsyncJobWorkersRunning sync.WaitGroup // 当前运行的异步工作者数量（用于等待全部退出）
)

// AsyncCallback 表示异步任务完成后，在游戏主线程中执行的回调函数，参数为结果与错误
type AsyncCallback func(res interface{}, err error)

func (ac AsyncCallback) callback(res interface{}, err error) { // 内部包装：在主线程中安全地触发回调
    if ac != nil {                                             // 回调不为空时才投递
        post.Post(func() { // 将回调函数投递到主游戏协程，避免并发访问游戏状态
            ac(res, err)
        })
    }
}

// AsyncRoutine 表示在独立 Goroutine 中执行的异步任务函数，返回结果与错误
type AsyncRoutine func() (res interface{}, err error)

type asyncJobWorker struct { // 异步作业工作者：管理一个任务队列并在后台循环处理
    jobQueue chan asyncJobItem // 任务队列，缓冲长度受 consts.ASYNC_JOB_QUEUE_MAXLEN 限制
}

type asyncJobItem struct { // 队列中单个任务项
    routine  AsyncRoutine  // 要执行的异步函数
    callback AsyncCallback // 完成后在主线程触发的回调
}

func newAsyncJobWorker() *asyncJobWorker { // 创建并启动一个新的异步工作者
    ajw := &asyncJobWorker{
        jobQueue: make(chan asyncJobItem, 10000), // 创建带缓冲的任务队列
    }
    numAsyncJobWorkersRunning.Add(1) // 记录有一个新的 worker 运行中
    go ajw.loop()                    // 启动后台处理循环
    return ajw                       // 返回工作者实例
}

func (ajw *asyncJobWorker) appendJob(routine AsyncRoutine, callback AsyncCallback) { // 追加一个任务到队列
    ajw.jobQueue <- asyncJobItem{routine, callback} // 入队，等待后台处理
}

func (ajw *asyncJobWorker) loop() { // 后台循环：持续从队列读取并执行任务
    defer numAsyncJobWorkersRunning.Done() // 循环退出时减少 WaitGroup 计数

    gwutils.RepeatUntilPanicless(func() { // 防护循环：若出现 panic 自动恢复并继续
        for item := range ajw.jobQueue {  // 逐个取出任务直到队列关闭
            res, err := item.routine()       // 在后台 Goroutine 中执行任务函数
            item.callback.callback(res, err) // 将结果投递到主线程并触发回调
        }
    })
}

var (
    asyncJobWorkersLock sync.RWMutex                   // 保护 worker 映射的读写锁
    asyncJobWorkers     = map[string]*asyncJobWorker{} // 按组名维护的异步工作者集合
)

func getAsyncJobWorker(group string) (ajw *asyncJobWorker) { // 获取指定分组的异步工作者（若不存在则懒创建）
    asyncJobWorkersLock.RLock()  // 先读锁快速路径
    ajw = asyncJobWorkers[group] // 查找已有的 worker
    asyncJobWorkersLock.RUnlock()

    if ajw == nil { // 双重检查：不存在则在写锁下创建
        asyncJobWorkersLock.Lock()
        ajw = asyncJobWorkers[group]
        if ajw == nil {
            ajw = newAsyncJobWorker()    // 创建新的 worker
            asyncJobWorkers[group] = ajw // 放入映射
        }
        asyncJobWorkersLock.Unlock()
    }
    return
}

// AppendAsyncJob 追加一个异步任务（在独立 Goroutine 执行，回调在游戏主线程触发）
func AppendAsyncJob(group string, routine AsyncRoutine, callback AsyncCallback) { // 将任务追加到指定分组的队列
    ajw := getAsyncJobWorker(group)  // 获取或创建对应的 worker
    ajw.appendJob(routine, callback) // 入队
}

// WaitClear 等待所有异步工作者退出（应仅在游戏主线程中调用）
func WaitClear() bool { // 关闭所有队列并阻塞直到工作者全部退出，返回是否进行了清理
    var cleared bool // 标记是否进行了清理操作
    // Close all job queue workers
    fmt.Printf("Waiting for all async job workers to be cleared ...") // 日志提示正在清理
    asyncJobWorkersLock.Lock()                                        // 写锁保护 worker 集合
    if len(asyncJobWorkers) > 0 {                                     // 存在正在运行的 worker
        for group, alw := range asyncJobWorkers { // 逐个关闭队列
            close(alw.jobQueue)             // 关闭队列，通知后台循环退出
            fmt.Printf("\tclear %s", group) // 输出清理的分组
        }
        asyncJobWorkers = map[string]*asyncJobWorker{} // 清空映射
        cleared = true                                 // 标记为已清理
    }
    asyncJobWorkersLock.Unlock() // 释放锁

    // wait for all job workers to quit
    numAsyncJobWorkersRunning.Wait() // 等待所有后台循环退出
    return cleared                   // 返回清理结果
}
