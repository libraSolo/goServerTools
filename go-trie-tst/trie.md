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
- `type Trie struct { Val interface{}; children map[byte]*Trie }`
- `func NewTrie() *Trie`：创建并返回一个新的根节点或子树节点。
- `func (t *Trie) Child(b byte) *Trie`：返回字节 `b` 的子节点；若不存在则懒创建。
- `func (t *Trie) ChildIfExists(b byte) *Trie`：返回字节 `b` 的子节点；若不存在则返回 `nil`（不创建）。
- `func (t *Trie) HasChild(b byte) bool`：判断在字节 `b` 上是否存在子节点。
- `func (t *Trie) Sub(s string) *Trie`：沿字符串 `s` 的每个字节逐层访问；路径上缺失节点则懒创建。
- `func (t *Trie) Find(s string) *Trie`：沿字符串 `s` 的每个字节逐层访问；路径上缺失节点则返回 `nil`（不创建）。
- `func (t *Trie) Keys() []byte`：返回当前节点所有子分支的字节列表（无序）。
- `func (t *Trie) DeleteChild(b byte)`：删除字节 `b` 的子节点（若不存在则忽略）。
- `func (t *Trie) Size() int`：统计从当前节点出发（包含自身）的节点总数。
- `func (t *Trie) Walk(visit func(path string, node *Trie) bool)`：从当前节点进行深度优先遍历；`path` 为累积路径；返回 `false` 可跳过继续深入该分支。
- `func (t *Trie) WalkFrom(prefix string, visit func(path string, node *Trie) bool)`：从指定前缀出发进行深度优先遍历；前缀不存在则不操作。

示例：
```go
var root trietst.Trie
node := root.Sub("foo") // f -> o -> o
node.Val = myValue       // 在该前缀节点挂载业务对象
```
 
## 进阶示例
- 查找且不创建（区分 Sub 与 Find）：
```go
var root trietst.Trie
root.Sub("ab").Val = 1

// Sub 会创建路径（若不存在）
_ = root.Sub("ac")

// Find 仅查找，不创建
if n := root.Find("ad"); n == nil {
    // "ad" 不存在
}
```

- 遍历全部路径或从前缀遍历：
```go
var root trietst.Trie
root.Sub("ab").Val = "X"
root.Sub("ac").Val = "Y"

// 全树遍历
root.Walk(func(path string, n *trietst.Trie) bool {
    // 处理 path 与 n.Val
    return true // 返回 false 跳过深入该分支
})

// 从前缀遍历
root.WalkFrom("a", func(path string, n *trietst.Trie) bool {
    return true
})
```

- 工具方法与监控：
```go
var root trietst.Trie
_ = root.Sub("ab")
_ = root.Sub("ac")

// 子分支键
bs := root.Child('a').Keys() // 可能为 ['b','c']，无序

// 删除子分支
root.Child('a').DeleteChild('c')

// 节点规模
total := root.Size()
```

## 与 pubsub 的协作
- `pubsub/pubsub/generic_pubsub.go`：
  - 通过 `Trie.Sub(prefix)` 定位前缀节点，在其 `Val` 挂载订阅集合。
  - 发布流程：
    - 逐层调用通配订阅者（`prefix + '*'`）。
    - 末端节点调用精确订阅者（完整主题）。

## 复杂度与代价
- 查找：`Sub(s)` / `Find(s)` 为 `O(len(s))`；`Child(b)` / `ChildIfExists(b)` 平均 `O(1)`。
- 空间：分支越多、路径越长，节点数与内存占用越高。
- 遍历：`Walk` / `WalkFrom` 为 `O(N)`（`N` 为从起点出发的节点数）。

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
- 路径级删除：当前支持删除单子分支（`DeleteChild`）；如需整条路径或子树清理，可引入父指针或路径栈。
- 遍历能力：已提供 `Walk`/`WalkFrom`，可按需扩展遍历顺序或过滤策略。
- 监控：节点计数、分支分布、热点路径统计。
- 序列化：持久化与恢复前缀树结构。

## 工作区集成与测试
- 根目录 `go.work` 包含 `use ./go-trie-tst`，本地模块可被其它模块优先解析。
- 测试建议：
  - 在 `go-trie-tst` 目录运行：`go test`
  - 或在仓库根运行：`go test ./go-trie-tst`
  - 若验证与 `pubsub` 的协作：在 `pubsub/pubsub` 目录运行：`go test` 或在仓库根运行：`go test ./pubsub/pubsub`