# Chart 排行榜服务（混合策略）

本模块提供基于“跳表 + 前 K 名最小堆 + 轻量缓存”的混合策略排行榜，实现高并发下的分数更新、精确排名查询和 TopN 读取。HTTP 接口使用 Gin 暴露，默认监听 `:8080`。

## 目录结构
```
chart/
├── api/               # 接口层（Gin 路由与处理器）
│   └── handle.go      # HTTP 路由注册与请求处理
├── domain/            # 领域层（核心数据结构与算法）
│   ├── cache.go       # RankCache：TopN 轻量缓存
│   ├── heap.go        # TopPlayersHeap：维护前 K 名
│   ├── leaderboard.go # HybridLeaderboard：混合排行榜聚合根
│   ├── player.go      # Player：玩家实体（Rank 仅用于响应填充）
│   └── skipList.go    # SkipList：精确排名（O(log n)）
├── storage/           # 基础设施层（仓储抽象与示例实现）
│   ├── repository.go  # 仓储接口定义
│   ├── memory.go      # 内存仓储示例（依赖工作区 rank-system/domain）
│   └── multiBackend.go# 多后端组合（示例/预留）
├── main.go            # 程序入口：初始化默认排行榜与路由
└── chart.md           # 本说明文档
```

## 架构概览
```
Client -> Gin Router -> API Handler -> HybridLeaderboard
                               ├── SkipList（精确排名）
                               ├── TopPlayersHeap（前 K 快速读取）
                               └── RankCache（TopN 短时缓存）
Storage：MemoryRepository（示例，可扩展为 Redis/SQL 等）
```

## 核心能力
- 更新分数：`PUT /api/v1/scores`（批量通道 + 同步回退）
- 查询玩家排名：`GET /api/v1/player-rank`（跳表精确排名，O(log n)）
- 查询前 N 名：`GET /api/v1/top-ranks`（缓存/跳表生成，近似 O(1)）
- 获取榜单信息：`GET /api/v1/leaderboard`

## HTTP 接口
- `PUT /api/v1/scores?leaderboard_id=<id>`
  - Body：`{ "player_id": number, "score": number }`
  - 返回：`{ "status": "success" }`
- `GET /api/v1/player-rank?leaderboard_id=<id>&player_id=<id>`
  - 返回：`{ "player_id": number, "rank": number }`
- `GET /api/v1/top-ranks?leaderboard_id=<id>&limit=<n>`
  - 返回：`[{ "id": number, "score": number, "rank": number, "update_time": string }, ...]`
- `GET /api/v1/leaderboard?leaderboard_id=<id>`
  - 返回：`{ "id": string, "name": string, "player_count": number, "config": {...} }`

## 关键设计与复杂度
- 跳表 SkipList：插入/删除/排名查询约 `O(log n)`；同分时按 `UpdateTime` 与 `ID` 稳定排序。
- 前 K 名 TopPlayersHeap：维护高分集，`Push/Pop O(log K)`，读取近似 `O(1)`。
- RankCache：以 `limit` 为键缓存 TopN，短 TTL（例如数秒）兼顾实时性与性能；返回副本避免竞态。
- 批量更新通道：生产者将更新写入 `batchUpdates`；通道满时自动回退到同步更新，降低丢包风险。
- 一致性：每次批处理后提升 `version` 并 `Invalidate()` 缓存；读取路径不修改共享实体。

## 运行与工作区说明
- 本模块已自包含，不再依赖 `rank-system/domain`。所有领域与存储类型均在 `chart/domain` 与 `chart/storage` 下实现。
- 入口 `main.go` 会创建默认榜单并注册路由：
  - 运行：`go run ./chart/chart`
  - 监听：`:8080`

## 注意事项
- `Player.Rank` 字段仅用作响应 DTO 填充，实体内的排名不持久存储；请通过接口或服务层实时计算排名。
- TopN 返回为副本，避免外部修改导致共享数据一致性问题。
- 大规模并发写入可通过批量通道实现，断言前需确保后台批处理完成（参见测试用例）。

## 扩展建议
- 增加持久化后端（Redis/SQL），实现 `storage.Repository` 的真实读写；
- 丰富查询接口（邻近排名、区间查询、分页 TopN）；
- 分季/分片策略与多榜单管理；
- 监控与指标（延迟、吞吐、缓存命中率）。

以上文档已与当前代码结构对齐，便于快速理解与集成使用。

