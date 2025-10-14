# 异步与主线程投递模块总结
**出自`https://github.com/xiaonanln/goworld``**
本项目的 async 目录下包含三个相互协作的模块：异步任务管理（async）、主线程投递（post）、容错工具（gwutils）。它们共同实现“在后台执行任务，并将结果安全地回投到主线程”的工作流。

## 目录结构
- async/async：核心异步任务管理，提供 AppendAsyncJob、WaitClear 等
- async/post：主线程回调投递与批处理执行，提供 Post、Tick
- async/gwutils：容错与工具函数，提供 CatchPanic、RunPanicless、RepeatUntilPanicless、NextLargerKey

## 模块职责与关键 API

1) async 模块（后台任务执行与分组管理）
- AsyncRoutine：在独立 Goroutine 执行的任务函数类型，形如 func() (res interface{}, err error)
- AsyncCallback：任务完成后在主线程执行的回调类型，形如 func(res interface{}, err error)
- AppendAsyncJob(group string, routine AsyncRoutine, callback AsyncCallback)：将任务追加到指定分组队列，routine 在后台执行，callback 在主线程触发
- WaitClear() bool：关闭所有任务队列并等待后台工作者退出
- 设计要点：
  - 每个分组对应一个后台 worker（asyncJobWorker），内部以 chan 作为任务队列
  - worker.loop 使用 gwutils.RepeatUntilPanicless 包裹循环，确保即便任务中发生 panic 也能恢复继续服务
  - callback 的触发通过 post.Post 投递到主线程，避免并发访问主线程状态

2) post 模块（主线程安全执行）
- Post(f PostCallback)：将无参回调投递到主线程执行队列
- Tick()：由主线程周期性调用，批量取出并以 gwutils.RunPanicless 执行回调
- 并发安全：队列由互斥锁保护，Tick 通过“复制当前队列再清空”的方式降低持锁时间

3) gwutils 模块（容错与辅助工具）
- CatchPanic(f func()) interface{}：执行函数并捕获 panic 返回错误信息
- RunPanicless(f func()) bool：安全执行函数，若发生 panic 则返回 false
- RepeatUntilPanicless(f func())：重复执行直到不再产生 panic，适用于“后台循环必须继续”的场景
- NextLargerKey(key string) string：生成严格大于给定字符串键的下一键（字典序），常用于范围查询的边界辅助

## 工作流程概览
1. 游戏逻辑在任意 Goroutine 发起 AppendAsyncJob，指定分组与 routine/callback。
2. 后台 worker 从队列取出 routine 执行，得到 res/err。
3. worker 将 callback 通过 post.Post 投递到主线程队列。
4. 主线程周期性调用 post.Tick，批量并安全地执行所有已投递的回调函数。
5. 如果需要退出或重启，调用 WaitClear 关闭所有队列并等待后台 worker 退出。

## 使用示例（简化）
- 主线程启动后，开启一个循环周期性调用 Tick：
  - go func() { for { post.Tick(); time.Sleep(time.Millisecond) } }()
- 提交任务并在主线程回调：
  - AppendAsyncJob("group1", func() (interface{}, error) { return 42, nil }, func(res interface{}, err error) { /* 在主线程处理结果 */ })

## 并发与健壮性
- 使用 sync.RWMutex 保护 worker 映射，确保并发获取或懒创建安全
- 通过 WaitGroup 记录后台 worker 存活数量，以便退出时等待清理
- 任务队列采用带缓冲 channel，避免过度阻塞；若业务需要可调整缓冲大小或增加限流策略
- 所有回调在主线程执行，避免数据竞争；回调执行过程使用 RunPanicless 防止单次失败影响批次

## 工作区（go.work）与本地依赖
- 根目录 go.work 已启用工作区并包含：async/async、async/post、async/gwutils
- 在根目录运行测试建议使用：
  - go test ./async/async ./async/post ./async/gwutils
  - 或分别进入模块目录执行 go test

## 常见问题与建议
- 问题：使用 go test ./... 可能提示“前缀不包含 go.work 列出的模块”，可改用上面的按模块路径测试方式
- 建议：
  - 为不同业务场景使用不同的分组名称，避免单队列拥塞
  - 在 routine 中自行处理超时与取消（如使用 context），提高资源回收能力
  - 根据负载监控动态调整队列长度或分组数量，避免内存占用过高