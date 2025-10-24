package timeWheel

import (
    "container/list"
    "errors"
    "log"
    "sync"
    "time"
)

// TimeWheel 多层时间轮结构定义
type TimeWheel struct {
    tickMs      time.Duration // 当前时间轮的基本时间跨度
    wheelSize   int           // 时间轮大小（槽位数）
    interval    time.Duration // 总时间跨度 tickMs * wheelSize
    currentTime int64         // 当前时间（毫秒时间戳）
    buckets     []*list.List  // 时间格队列

    queue     *DelayQueue  // 延迟队列，用于精准推进
    taskMutex sync.RWMutex // 任务锁

    // 层级相关
    overflowWheel *TimeWheel // 上层时间轮
    lowerLevel    *TimeWheel // 下层时间轮（用于降级）
    level         int        // 当前层级

    // 运行控制
    stopCh  chan struct{}
    running bool
}

// TimerTask 定时任务
type TimerTask struct {
    delay    time.Duration // 任务延迟时间
    key      interface{}   // 任务唯一标识
    job      func()        // 任务执行函数
    deadline int64         // 绝对到期时间（毫秒）
    level    int           // 任务当前所在层级
}

// DelayQueue 延迟队列
type DelayQueue struct {
    C        chan interface{}
    pq       priorityQueue
    mutex    sync.Mutex
    sleeping int32
    wakeupCh chan struct{}
}

// 优先级队列项
type item struct {
    Value    interface{}
    Priority int64 // 优先级（到期时间）
    Index    int
}

// priorityQueue 小根堆实现的优先级队列
type priorityQueue []*item

func (pq priorityQueue) Len() int           { return len(pq) }
func (pq priorityQueue) Less(i, j int) bool { return pq[i].Priority < pq[j].Priority }
func (pq priorityQueue) Swap(i, j int) {
    pq[i], pq[j] = pq[j], pq[i]
    pq[i].Index = i
    pq[j].Index = j
}

func (pq *priorityQueue) Push(x interface{}) {
    n := len(*pq)
    c := cap(*pq)
    if n+1 > c {
        npq := make(priorityQueue, n, c*2)
        copy(npq, *pq)
        *pq = npq
    }
    *pq = (*pq)[0 : n+1]
    item := x.(*item)
    item.Index = n
    (*pq)[n] = item
}

func (pq *priorityQueue) Pop() interface{} {
    n := len(*pq)
    c := cap(*pq)
    if n < (c/2) && c > 25 {
        npq := make(priorityQueue, n, c/2)
        copy(npq, *pq)
        *pq = npq
    }
    item := (*pq)[n-1]
    item.Index = -1
    *pq = (*pq)[0 : n-1]
    return item
}

// PeekAndShift 查看并弹出到期元素
func (pq *priorityQueue) PeekAndShift(max int64) (*item, int64) {
    if pq.Len() == 0 {
        return nil, 0
    }

    item := (*pq)[0]
    if item.Priority > max {
        return nil, item.Priority - max
    }
    old := *pq
    n := len(old)
    *pq = old[1:n]
    for i := 0; i < len(*pq); i++ {
        (*pq)[i].Index = i
    }
    return item, 0
}

// NewDelayQueue 创建延迟队列
func NewDelayQueue(size int) *DelayQueue {
    return &DelayQueue{
        C:        make(chan interface{}),
        pq:       make(priorityQueue, 0, size),
        wakeupCh: make(chan struct{}),
    }
}

// Offer 写入元素到延迟队列
func (dq *DelayQueue) Offer(elem interface{}, expiration int64) {
    item := &item{
        Value:    elem,
        Priority: expiration,
    }

    dq.mutex.Lock()
    dq.pq = append(dq.pq, item)
    for i := len(dq.pq)/2 - 1; i >= 0; i-- {
        dq.down(i)
    }
    dq.mutex.Unlock()

    select {
    case dq.wakeupCh <- struct{}{}:
    default:
    }
}

// Poll 从延迟队列中获取元素
func (dq *DelayQueue) Poll(exitC chan struct{}, nowF func() int64) {
    for {
        now := nowF()

        dq.mutex.Lock()
        item, delta := dq.pq.PeekAndShift(now)
        dq.mutex.Unlock()

        if item != nil {
            select {
            case dq.C <- item.Value:
            case <-exitC:
                return
            }
        } else {
            if delta == 0 {
                select {
                case <-dq.wakeupCh:
                    continue
                case <-exitC:
                    return
                }
            } else if delta > 0 {
                timer := time.NewTimer(time.Duration(delta) * time.Millisecond)
                select {
                case <-dq.wakeupCh:
                    timer.Stop()
                    continue
                case <-timer.C:
                    continue
                case <-exitC:
                    timer.Stop()
                    return
                }
            }
        }
    }
}

// down 堆的下沉操作
func (dq *DelayQueue) down(i int) {
    n := len(dq.pq)
    for {
        left := 2*i + 1
        if left >= n {
            break
        }
        smallest := left
        if right := left + 1; right < n && dq.pq.Less(right, left) {
            smallest = right
        }
        if dq.pq.Less(i, smallest) {
            break
        }
        dq.pq.Swap(i, smallest)
        i = smallest
    }
}

// New 创建时间轮实例
func New(tickMs time.Duration, wheelSize int) (*TimeWheel, error) {
    if tickMs <= 0 || wheelSize <= 0 {
        return nil, errors.New("tickMs and wheelSize must be greater than zero")
    }

    tw := &TimeWheel{
        tickMs:    tickMs,
        wheelSize: wheelSize,
        interval:  tickMs * time.Duration(wheelSize),
        buckets:   make([]*list.List, wheelSize),
        queue:     NewDelayQueue(wheelSize),
        stopCh:    make(chan struct{}),
        level:     1,
    }

    for i := 0; i < wheelSize; i++ {
        tw.buckets[i] = list.New()
    }

    return tw, nil
}

// Start 启动时间轮
func (tw *TimeWheel) Start() {
    tw.taskMutex.Lock()
    defer tw.taskMutex.Unlock()

    if tw.running {
        return
    }
    tw.running = true
    tw.currentTime = truncateToTick(timeToMs(time.Now()), int64(tw.tickMs/time.Millisecond))

    go func() {
        tw.queue.Poll(tw.stopCh, func() int64 {
            return timeToMs(time.Now())
        })
    }()

    go tw.processQueue()

    log.Printf("时间轮启动，层级: %d, 间隔: %v", tw.level, tw.tickMs)
}

// Stop 停止时间轮
func (tw *TimeWheel) Stop() {
    tw.taskMutex.Lock()
    defer tw.taskMutex.Unlock()

    if !tw.running {
        return
    }

    tw.running = false
    close(tw.stopCh)

    if tw.overflowWheel != nil {
        tw.overflowWheel.Stop()
    }

    log.Println("时间轮已停止")
}

// processQueue 处理延迟队列
func (tw *TimeWheel) processQueue() {
    for {
        select {
        case bucket := <-tw.queue.C:
            tw.expireBucket(bucket.(*list.List))
        case <-tw.stopCh:
            return
        }
    }
}

// expireBucket 处理到期的时间格
func (tw *TimeWheel) expireBucket(bucket *list.List) {
    tw.taskMutex.Lock()
    defer tw.taskMutex.Unlock()

    currentMs := timeToMs(time.Now())
    tw.currentTime = currentMs

    for e := bucket.Front(); e != nil; {
        next := e.Next()
        task := e.Value.(*TimerTask)
        bucket.Remove(e)

        // 检查任务是否真的到期
        if task.deadline <= currentMs {
            // 任务到期，执行
            go func(t *TimerTask) {
                t.job()
                log.Printf("任务执行: key=%v, 实际执行层级=%d", t.key, tw.level)
            }(task)
        } else {
            // 任务未真正到期，需要降级或重新安排
            if tw.lowerLevel != nil {
                // 有下层时间轮，进行降级
                // 更新任务的层级信息
                task.level = tw.lowerLevel.level
                tw.lowerLevel.addTaskInternal(task, currentMs)
                log.Printf("任务降级: key=%v, 从层级%d降级到层级%d", task.key, tw.level, tw.lowerLevel.level)
            } else {
                // 没有下层时间轮（最底层），重新计算并添加
                tw.addTaskInternal(task, currentMs)
                log.Printf("任务重新安排: key=%v, 层级=%d", task.key, tw.level)
            }
        }
        e = next
    }
}

// AddTask 添加定时任务
func (tw *TimeWheel) AddTask(delay time.Duration, key interface{}, job func()) error {
    if delay < 0 {
        return errors.New("delay must be non-negative")
    }

    task := &TimerTask{
        delay:    delay,
        key:      key,
        job:      job,
        deadline: timeToMs(time.Now().Add(delay)),
        level:    tw.level, // 记录任务的初始层级
    }

    tw.taskMutex.Lock()
    defer tw.taskMutex.Unlock()

    return tw.addTaskInternal(task, timeToMs(time.Now()))
}

// addTaskInternal 内部添加任务方法
func (tw *TimeWheel) addTaskInternal(task *TimerTask, currentMs int64) error {
    // 如果任务已到期，立即执行
    if task.deadline <= currentMs {
        go func(t *TimerTask) {
            t.job()
            log.Printf("任务立即执行: key=%v, 层级=%d", t.key, t.level)
        }(task)
        return nil
    }

    // 计算相对延迟
    relativeDelay := time.Duration(task.deadline-currentMs) * time.Millisecond

    // 如果任务超出当前时间轮范围，添加到上层时间轮
    if relativeDelay >= tw.interval {
        if tw.overflowWheel == nil {
            // 懒加载创建上层时间轮
            overflowWheel, err := New(tw.interval, tw.wheelSize)
            if err != nil {
                return err
            }
            overflowWheel.level = tw.level + 1
            overflowWheel.currentTime = currentMs
            overflowWheel.lowerLevel = tw // 设置下层时间轮引用
            tw.overflowWheel = overflowWheel

            // 启动上层时间轮
            if tw.running {
                tw.overflowWheel.Start()
            }
        }
        // 更新任务的层级信息
        task.level = tw.overflowWheel.level
        return tw.overflowWheel.addTaskInternal(task, currentMs)
    }

    // 计算在当前时间轮中的位置
    ticks := int(relativeDelay / tw.tickMs)
    if ticks < 1 {
        ticks = 1
    }

    // 修正位置计算
    currentTick := currentMs / int64(tw.tickMs/time.Millisecond)
    bucketIndex := int((currentTick + int64(ticks)) % int64(tw.wheelSize))

    targetBucket := tw.buckets[bucketIndex]

    // 添加到时间格
    targetBucket.PushBack(task)

    // 计算时间格的到期时间
    bucketExpiration := currentMs + int64(ticks)*int64(tw.tickMs/time.Millisecond)

    // 将时间格添加到延迟队列
    tw.queue.Offer(targetBucket, bucketExpiration)

    log.Printf("添加任务: key=%v, 延迟=%v, 层级=%d, 位置=%d, 到期时间=%v",
        task.key, task.delay, tw.level, bucketIndex, time.Unix(0, task.deadline*int64(time.Millisecond)))
    return nil
}

// RemoveTask 删除任务
func (tw *TimeWheel) RemoveTask(key interface{}) bool {
    tw.taskMutex.Lock()
    defer tw.taskMutex.Unlock()

    // 遍历所有时间格查找并删除任务
    for i, bucket := range tw.buckets {
        for e := bucket.Front(); e != nil; e = e.Next() {
            if task, ok := e.Value.(*TimerTask); ok && task.key == key {
                bucket.Remove(e)
                log.Printf("删除任务: key=%v, 层级=%d, 位置=%d", key, tw.level, i)
                return true
            }
        }
    }

    // 在上层时间轮中查找
    if tw.overflowWheel != nil {
        return tw.overflowWheel.RemoveTask(key)
    }

    return false
}

// timeToMs 时间转毫秒
func timeToMs(t time.Time) int64 {
    return t.UnixNano() / int64(time.Millisecond)
}

// msToTime 毫秒转时间
func msToTime(ms int64) time.Time {
    return time.Unix(0, ms*int64(time.Millisecond))
}

// truncateToTick 将时间截断到tick的整数倍
func truncateToTick(ms, tickMs int64) int64 {
    if tickMs <= 0 {
        return ms
    }
    return ms - (ms % tickMs)
}
