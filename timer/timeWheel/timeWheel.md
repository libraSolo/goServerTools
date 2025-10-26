# 时间轮：高效调度百万定时任务的巧妙算法（Go 实现详解）

> 高效调度，事半功倍

本文结合包 `timeWheel` 的源码实现（已拆分为多个模块），从概念、工程设计到使用示例，系统性讲清如何用时间轮在 Go 中构建一个高并发、可取消的定时调度引擎。

---

## 背景与动机

电商订单超时、连接心跳、延迟消息和批量任务调度等场景，常常需要同时管理海量定时任务，并要求：
- 插入/删除高效；
- 精准推进，避免空转；
- 支持取消，且并发安全。

传统解法局限：
- `Timer`：串行、易受慢任务影响；
- `ScheduledThreadPool`：并发提升有限，大量任务时性能下滑；
- 仅用 `DelayQueue`：插入/删除 `O(log n)`，难以承载层级降级与批量处理需求。

时间轮（TimingWheel）通过“环形复用 + 层级降级 + 精准推进”，在性能与功能之间取得平衡，已被 Kafka、ZooKeeper 等广泛采用。

---

## 核心概念（与时钟类比）

- `tick`：每个时间格的跨度（如 `100ms`）。
- `wheelSize`：时间格数量（如 `512`）。
- `interval`：单轮覆盖范围，`tick * wheelSize`。
- `currentTime`：当前指针（对齐到 `tick` 的整数倍）。

任务被放入对应时间格；指针推进到该格时，任务出格并“降级/重插”，直到真正到期执行。

---

## 模块拆分与职责

源码按功能拆分为 5 个文件，职责清晰、便于维护：
- `priority_queue.go`：最小堆优先队列，按到期时间排序元素。
- `delay_queue.go`：延时队列，负责精准推进与到期投递（避免空转）。
- `bucket.go`：时间格，批量存放同一到期片段的任务，过期时统一处理。
- `timer_task.go`：任务实体，支持取消（`Stop`）与并发安全管理。
- `timewheel.go`：时间轮核心，层级溢出、时间推进与任务降级逻辑。

---

## 快速上手（同包示例）

以下示例在包 `timeWheel` 内调用；若需包外使用，可封装导出方法（例如 `Schedule`）。

```go
now := time.Now().UnixNano() / 1e6

// 1) 初始化并启动
dq := NewDelayQueue(64)
tw := NewTimeWheel(100, 512, now, dq) // tick=100ms, 512个格
tw.Start()

// 2) 添加一个 500ms 后执行的任务
t := &TimerTaskEntity{DelayTime: now + 500, Task: func(){ fmt.Println("run") }}
tw.tryAdd(t) // 在包内使用；包外可封装导出方法

// 3) 可选：取消任务（若尚未执行且仍在桶中）
_ = t.Stop()

// 4) 结束（确保后台循环退出）
tw.Stop()
```

---

## 设计亮点与实现要点

### 1）精准推进：空间换时间

- 每个到期的 `Bucket` 放入 `DelayQueue`；队头即最早到期元素。
- 只在队头到期时推进时间并处理任务，避免固定频率推进造成的空转。

`DelayQueue.Poll` 的实现已优化为更符合 Go 习惯的写法：使用 `defer` 做统一清理，并在各分支 `return`，移除 `goto`。

```go
func (dq *DelayQueue) Poll(exitC chan struct{}, nowF func() int64) {
    defer atomic.StoreInt32(&dq.sleeping, 0)
    for {
        now := nowF()
        dq.mu.Lock()
        item, delta := dq.pq.PeekAndShift(now)
        if item == nil { atomic.StoreInt32(&dq.sleeping, 1) }
        dq.mu.Unlock()

        if item == nil {
            if delta == 0 {
                select {
                case <-dq.wakeupC:
                    continue
                case <-exitC:
                    return
                }
            } else if delta > 0 {
                select {
                case <-dq.wakeupC:
                    continue
                case <-time.After(time.Duration(delta) * time.Millisecond):
                    if atomic.SwapInt32(&dq.sleeping, 0) == 0 { <-dq.wakeupC }
                    continue
                case <-exitC:
                    return
                }
            }
        }

        select {
        case dq.C <- item.Value:
        case <-exitC:
            return
        }
    }
}
```

### 2）层级时间轮：承载更长延时

- 当前轮无法容纳的任务（超出 `interval`），溢出至上层时间轮（按需创建）。
- 到期后按剩余时间逐层降级，最终在最底层执行。

```go
func (tw *TimeWheel) add(t *TimerTaskEntity) bool {
    currentTime := atomic.LoadInt64(&tw.currentTime)
    if t.DelayTime < currentTime+tw.tick {
        return false // 已到执行窗口，直接执行
    } else if t.DelayTime < currentTime+tw.interval {
        virtualID := t.DelayTime / tw.tick
        bucket := tw.buckets[virtualID%tw.wheelSize]
        bucket.Add(t)
        if bucket.SetExpiration(virtualID * tw.tick) { tw.queue.Offer(bucket, bucket.Expiration()) }
        return true
    } else {
        if tw.overflow == nil {
            tw.overflow = NewTimeWheel(tw.interval, tw.wheelSize, currentTime, tw.queue)
        }
        return tw.overflow.add(t)
    }
}
```

### 3）并发安全与可取消

- `TimerTaskEntity` 使用原子指针记录所在桶，`Stop()` 在任务未执行前可取消：

```go
func (t *TimerTaskEntity) Stop() bool {
    stopped := false
    for b := t.getBucket(); b != nil; b = t.getBucket() { stopped = b.Remove(t) }
    return stopped
}
```

- `Bucket.Flush()` 采用“先集中移除再回调重插”，避免对 `Remove` 的重入导致死锁，并减少长时间持锁对其他操作的影响：

```go
func (b *Bucket) Flush(reinsert func(t *TimerTaskEntity)) {
    b.mu.Lock()
    var ts []*TimerTaskEntity
    for e := b.tasks.Front(); e != nil; e = e.Next() {
        t := e.Value.(*TimerTaskEntity)
        b.tasks.Remove(e)
        t.setBucket(nil)
        t.element = nil
        ts = append(ts, t)
    }
    b.SetExpiration(-1)
    b.mu.Unlock()
    for _, t := range ts { reinsert(t) }
}
```

---

## 参数选型建议

- `tick`（时间格跨度）：越小精度越高，但格子越多、开销越大。一般 `50–200ms` 为工程上合适的折中。
- `wheelSize`（格子数）：影响单轮覆盖范围与桶数量，常见取值 `512/1024`。
- `DelayQueue` 初始容量：根据期望并发与到期分布设置，例如 `64/128`。

---

## 适用场景与注意事项

- 适用：心跳检测、连接超时、延迟消息、批量定时调度。
- 精度：由 `tick` 决定；例如 `tick=100ms` 时，不适合处理亚 100ms 的任务。
- 内存与性能：时间格与任务列表为常驻结构，适合“任务量大、到期分布广”的场景。
- 取消语义：`Stop()` 仅保证“尚未执行时可取消”；已出格或正在执行的任务可能无法取消。

---

## 总结

时间轮以“环形复用 + 层级降级 + 精准推进”的组合策略，在海量定时任务场景中兼顾了性能与工程可维护性。当前实现进行了模块化拆分，并通过并发安全与死锁修正（`Flush` 设计、`Poll` 无 `goto`）增强了稳定性。

当你需要实现一个可扩展的定时调度引擎时，不妨试试时间轮：简单、可靠、可扩展。