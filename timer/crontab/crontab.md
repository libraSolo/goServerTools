# Go 实现的轻量级 Crontab 调度器

本项目是 `crontab` 包的详细介绍，这是一个纯 Go 实现的、轻量级的、进程内（in-process）定时任务调度器。它的设计目标是提供一个简单、高效、易于集成的定时任务解决方案，无需外部依赖。

## 核心设计

`crontab` 调度器主要由以下几个部分构成：

1.  **调度器 (`Scheduler`)**：作为核心引擎，`Scheduler` 内部维护了一个基于最小堆（min-heap）的定时器，用于高效管理所有待执行的任务。
2.  **任务实体 (`entry`)**：每个定时任务都被抽象为一个 `entry` 对象，包含了任务的调度规则（crontab 表达式）、下一次执行时间以及要执行的回调函数。
3.  **任务句柄 (`Handle`)**：每次成功注册一个任务后，会返回一个唯一的 `Handle`。该句柄可用于后续取消该任务。

其工作流程如下：
-   通过 `Initialize()` 启动一个全局的后台 Goroutine，该 Goroutine 每分钟被唤醒一次。
-   唤醒后，它会调用 `check(time.Now())`，遍历所有已注册的任务。
-   `check` 函数会判断当前时间是否满足任务的 crontab 表达式，如果满足，则在独立的 Goroutine 中异步执行该任务的回调函数。

## API 使用指南

### `Initialize()`
启动 `crontab` 调度器的后台服务。在应用启动时，应首先调用此函数。

```go
func Initialize()
```

### `Register(minute, hour, day, month, dayofweek int, cb func()) Handle`
注册一个新的定时任务。

-   `minute`, `hour`, `day`, `month`, `dayofweek`: 分别代表分钟、小时、日、月、星期。每个参数都可以是：
    -   一个非负数，表示一个确切的时间点（例如 `minute=30` 表示在 30 分时）。
    -   一个负数，表示一个时间间隔（例如 `minute=-5` 表示每 5 分钟）。
    -   `dayofweek` 中，0 和 7 都代表周日。
-   `cb`: 任务触发时要执行的回调函数。
-   返回值：成功则返回任务句柄 `Handle`。

```go
// 示例：注册一个每分钟执行一次的任务
// -1, -1, -1, -1, -1 分别对应 crontab 中的 * * * *
handle := Register(-1, -1, -1, -1, -1, func() {
    fmt.Println("Task is running!")
})
```

### `Unregister(handle Handle)`
根据提供的 `Handle` 取消一个已注册的定时任务。

```go
// 示例：取消之前注册的任务
Unregister(handle)
```

## 完整使用示例

下面是一个完整的使用示例，展示了如何初始化调度器、注册任务、等待任务执行，最后再取消任务。

```go
package main

import (
	"fmt"
	"time"
	"e.com/bro/goTools/timer/crontab"
)

func main() {
	// 1. 初始化调度器
	crontab.Initialize()
	fmt.Println("Crontab scheduler initialized.")

	var triggered bool
	// 2. 注册一个每分钟执行一次的任务
	handle := crontab.Register(-1, -1, -1, -1, -1, func() {
		fmt.Println("Callback triggered!")
		triggered = true
	})
	fmt.Printf("Task registered with handle: %d\n", handle)

	// 3. 等待足够长的时间以便任务至少执行一次
    // 在实际应用中，这里通常是主程序的事件循环
	fmt.Println("Waiting for task to be triggered...")
	time.Sleep(65 * time.Second)

	// 4. 验证任务是否被触发
	if triggered {
		fmt.Println("Test validation: SUCCESS")
	} else {
		fmt.Println("Test validation: FAILED")
	}

	// 5. 取消任务
	crontab.Unregister(handle)
	fmt.Println("Task unregistered.")
}
```

## 并发与安全

-   **回调执行**：每个任务的回调函数都在一个独立的 Goroutine 中执行，这保证了回调的执行不会阻塞调度器的主循环。
-   **线程安全**：用户需要自行确保回调函数内部的逻辑是线程安全的，尤其是在访问共享状态时。
-   **错误处理**：`Register` 函数会返回详细的解析错误，便于调试。回调函数中的 `panic` 会被捕获，但建议在回调函数内部做好错误处理。

## 测试策略

为了确保测试的稳定性和确定性，`crontab` 包的测试采用了以下策略：
-   `check(t time.Time)` 函数接受一个时间参数，允许测试代码注入一个模拟的“当前时间”，从而可以精确地验证在特定时间点哪些任务应该被触发。
-   测试用例避免使用 `time.Sleep` 来等待任务执行，而是直接调用 `check` 函数，使测试过程完全同步，消除了不确定性。
-   提供了 `reset()` 函数，用于在每次测试后清理全局状态，确保测试用例之间相互独立。