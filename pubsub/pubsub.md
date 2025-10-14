# PubSub（发布订阅）模块知识总结

本模块提供与引擎解耦的通用发布/订阅能力，支持“精确订阅”和“前缀通配订阅（仅允许末尾 '*'）”。适用于将事件或消息按主题路由到不同订阅者的场景。

## 模块结构
- `pubsub/pubsub/generic_pubsub.go`：核心发布订阅实现
- `pubsub/common/`：通用集合类型与工具（如 `StringSet`）

## 核心类型与 API
- `type Handler func(subject string, content string)`：订阅者回调函数类型
- `type GenericPubSub struct { ... }`：发布订阅服务实例
- `func NewGenericPubSub() *GenericPubSub`：创建服务实例
- `func (ps *GenericPubSub) Subscribe(subscriberID, subject string, handler Handler)`：订阅主题
  - 规则：`'*'` 仅允许在主题末尾且最多出现一次；`"*"` 表示订阅所有主题（任意前缀）
  - 同一订阅者多次订阅会更新其 `Handler`
- `func (ps *GenericPubSub) Unsubscribe(subscriberID, subject string)`：取消订阅（支持末尾通配）
- `func (ps *GenericPubSub) UnsubscribeAll(subscriberID string)`：取消该订阅者的所有订阅（精确与通配）
- `func (ps *GenericPubSub) Publish(subject, content string)`：发布主题与内容（主题中不允许出现 `'*'`）

## 前缀通配的工作原理
- 订阅阶段：
  - 精确订阅者存储在对应叶子节点的 `subscribers` 集合
  - 通配订阅者存储在每一层的 `wildcardSubscribers` 集合（对应 `prefix + '*'`）
- 发布阶段：
  - 沿前缀树从根到叶逐层触发通配订阅者（覆盖所有前缀匹配）
  - 在叶子节点触发精确订阅者（完整主题匹配）

## 依赖与实现细节
- 前缀树：通过本地 `go-trie-tst` 模块的 `TrieMO` 前缀树实现，支持：
  - `Sub(subjectPrefix)`：下钻到某前缀节点
  - `Child(byte)`：访问/创建子节点
  - `Val`：在树节点上挂载订阅集合
- 并发安全：
  - 使用 `sync.RWMutex` 保护订阅结构与回调映射
  - 发布阶段采用读锁，订阅与取消订阅阶段采用写锁

## 使用示例
```go
ps := pubsub.NewGenericPubSub()

// 订阅者 A：精确订阅 "news/sports"
ps.Subscribe("A", "news/sports", func(subject, content string) {
    println("A recv", subject, content)
})

// 订阅者 B：通配订阅 "news/*"（所有 news 前缀）
ps.Subscribe("B", "news/*", func(subject, content string) {
    println("B recv", subject, content)
})

// 发布：同时命中 B 的通配与 A 的精确订阅
ps.Publish("news/sports", "UCL results")

// 取消订阅示例
ps.Unsubscribe("A", "news/sports")
ps.UnsubscribeAll("B")
```

## 复杂度与性能
- 发布：沿主题逐字节下钻，时间复杂度约为 `O(len(subject) + 匹配订阅者数量)`
- 订阅/取消订阅：逐字节下钻并更新集合，复杂度约为 `O(len(subject))`
- 空间：前缀树节点与集合随主题数量增长；合理的前缀规划可降低碎片化

## 最佳实践
- 前缀规划：对主题命名进行分层规划（如 `domain/category/item`），便于通配订阅与边界控制
- 订阅者标识：`subscriberID` 应保持全局唯一，方便精准取消订阅
- 回调健壮性：在回调处理内做好错误兜底，避免影响其它订阅者执行
- 发布校验：发布主题不能包含 `'*'`；否则应先清洗或拒绝

## 工作区与本地依赖
- 根目录 `go.work` 已包含：`use ./pubsub/common`、`use ./pubsub/pubsub`、`use ./go-trie-tst`
- 本模块依赖会优先解析到本地实现，便于联调与修改

## 测试建议
- 在仓库根运行：`go test ./pubsub/pubsub`
- 或进入目录运行：`cd pubsub/pubsub && go test`