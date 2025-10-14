# Trie（前缀树）

本文档梳理本仓库 `go-trie-tst` 模块的设计、用途与最佳实践，并与常见前缀结构进行对比。

- 模块位置：`go-trie-tst/`
- 主要代码：`go-trie-tst/trie.go`
- 主要用途：为 `pubsub` 提供高效的前缀匹配能力（主题前缀 + 通配）。

## 设计目标
- 轻量、易用的前缀树（Trie）结构，支持按字节路径下钻与懒创建。
- 节点承载任意业务值（`Val`），由使用方自定义类型并断言。
- 与工作区（`go.work`）集成，供本地其它模块直接引用。

## 数据结构与 API
- `type TrieMO struct { Val interface{}; children map[byte]*TrieMO }`
- `func (t *TrieMO) Child(b byte) *TrieMO`
  - 返回字节 `b` 的子节点；若不存在则懒创建。
- `func (t *TrieMO) Sub(s string) *TrieMO`
  - 沿字符串 `s` 的每个字节逐层访问；路径上缺失节点则懒创建。

示例：
```go
var root trietst.TrieMO
node := root.Sub("foo") // f -> o -> o
node.Val = myValue       // 在该前缀节点挂载业务对象
```

## 与 pubsub 的协作
- `pubsub/pubsub/generic_pubsub.go`：
  - 通过 `TrieMO.Sub(prefix)` 定位前缀节点，在其 `Val` 挂载订阅集合。
  - 发布流程：
    - 逐层调用通配订阅者（`prefix + '*'`）。
    - 末端节点调用精确订阅者（完整主题）。

## 复杂度与代价
- 查找：`Sub(s)` 为 `O(len(s))`；`Child(b)` 平均 `O(1)`。
- 空间：分支越多、路径越长，节点数与内存占用越高。

## 与其他结构的对比
- 标准 Trie：逐字符/字节分支；适合前缀匹配与词典类问题。
- TST（三叉搜索树）：有序符号表风格，空间较省但实现更复杂。
- Radix/压缩 Trie：合并连续边，降低节点数；实现更复杂。
- 本实现采用“按字节的哈希分支”，优点是实现简单、查找快；缺点是极端分支时内存较高。

## 最佳实践
- 字符编码：按字节处理字符串。若需字符级语义（UTF-8），可自行用 `rune` 迭代并调整结构。
- 并发：节点不加锁；由上层（如 `pubsub`）用 `RWMutex` 控制并发。
- 类型安全：`Val` 为 `interface{}`，使用方需自行断言并保证一致。
- 资源管理：大量唯一前缀可能造成节点膨胀；可通过分组、压缩前缀或限流优化。

## 可扩展方向
- 删除与子树清理：提供移除路径/子树的 API（需路径栈或父指针）。
- 遍历：前序/深度遍历导出所有路径与挂载值。
- 监控：节点计数、分支分布、热点路径统计。
- 序列化：持久化与恢复前缀树结构。

## 工作区集成与测试
- 根目录 `go.work` 包含 `use ./go-trie-tst`，本地模块可被其它模块优先解析。
- 测试建议：
  - 在 `pubsub/pubsub` 目录运行：`go test`
  - 或在仓库根运行：`go test ./pubsub/pubsub`