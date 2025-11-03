# Chart 排行榜服务（混合策略）

本包实现基于“跳表 + 前K名最小堆 + 轻量缓存”的混合策略排行榜，提供高并发下的快速分数更新、精确排名查询和前 N 名查询。HTTP 接口基于 Gin 暴露，默认启动在 `:8080`。

## 目录结构
```
chart/
├── api/              # 接口层
│   └── handle.go     # 路由注册与 HTTP 处理器
├── domain/           # 领域层（核心数据结构与算法）
│   ├── cache.go      # 前 N 名结果缓存
│   ├── heap.go       # TopPlayersHeap 最小堆（维护前 K）
│   ├── leaderboard.go# HybridLeaderboard 混合排行榜
│   ├── player.go     # 玩家实体
│   └── skipList.go   # 跳表实现（精确排名）
├── storage/          # 基础设施层
│   ├── memory.go     # 内存仓储实现
│   ├── multiBackend.go# 预留：多后端组合
│   └── repository.go # 仓储接口定义
├── main.go           # 程序入口（初始化默认排行榜与路由）
└── chart.md          # 本说明文档
```

## 架构概览
```
┌─────────────────┐    ┌──────────────────┐    ┌───────────────────────────┐
│     Client      │ -> │     Gin Router    │ -> │        Handler (API)       │
└─────────────────┘    └──────────────────┘    └─────────────┬─────────────┘
                                                              │
                                         ┌────────────────────┴────────────────────┐
                                         │              Domain Layer               │
                                         │  HybridLeaderboard (跳表 + 最小堆 + 缓存) │
                                         └─────────────┬──────────────┬───────────┘
                                                       │              │
                                           ┌───────────┘              └───────────┐
                                           │                                       │
                                    SkipList (排名)                        TopPlayersHeap (前K)
                                           │                                       │
                                           └───────────────┬───────────────────────┘
                                                           │
                                                    RankCache (轻量缓存)

                                          ┌───────────────────────────────────────┐
                                          │            Storage Layer               │
                                          │        MemoryRepository（示例）        │
                                          └───────────────────────────────────────┘
```

## 核心能力
- 更新分数：`PUT /api/v1/scores`（支持高并发与批量通道）
- 查询玩家排名：`GET /api/v1/player-rank`（跳表 O(log n) 精确排名）
- 查询前 N 名：`GET /api/v1/top-ranks`（基于最小堆与缓存，近似 O(1)）
- 获取排行榜信息：`GET /api/v1/leaderboard`

## HTTP 接口
- `PUT /api/v1/scores?leaderboard_id=<id>`
  - Body：`{ "player_id": number, "score": number }`
  - 结果：`{ "status": "success" }`

- `GET /api/v1/player-rank?leaderboard_id=<id>&player_id=<id>`
  - 结果：`{ "player_id": number, "rank": number }`

- `GET /api/v1/top-ranks?leaderboard_id=<id>&limit=<n>`
  - 结果：`[ { "id": number, "score": number, "rank": number }, ... ]`

- `GET /api/v1/leaderboard?leaderboard_id=<id>`
  - 结果：`{ "id": string, "name": string, "player_count": number, "config": {...} }`

## 数据结构与复杂度
- 跳表（SkipList）：插入/删除/排名查询约 `O(log n)`；按更新时间与分数稳定排序
- 前 K 名最小堆（TopPlayersHeap）：维护高分集，`Push/Pop O(log K)`，读取近似 `O(1)`
- 结果缓存（RankCache）：针对不同 `limit` 缓存前 N 名，TTL 短（例如 2s）以兼顾实时性与性能
- 批量更新通道：后台聚合处理，减少锁竞争并提升吞吐

## 设计要点
- 精确排名与热点读取分离：跳表负责全量排名，堆与缓存负责热点 TopN
- 版本与失效：批量更新推进版本，主动失效缓存确保数据一致性
- 并发安全：读写锁保护玩家映射与堆结构；缓存读写分离

## 运行与工作区说明
- 本包的导入路径与模块设置依赖工作区（`go.work`）中 `chart/rank-system` 等模块；若需直接运行，请确保模块路径与 `go.mod` 配置完整。
- 典型入口：`main.go` 会创建一个默认排行榜并注册所有路由，然后在 `:8080` 启动服务。

## 后续扩展建议
- 增加持久化后端（如 Redis / SQL），完善 `storage.Repository` 的实现
- 丰富排行榜策略（分段、赛季、去重、多维度排名）
- 增加邻近排名、区间查询等接口（`GetNearbyRanks` 已具备基础能力）

——
以上内容对齐仓库内其他模块说明风格，便于快速理解与集成。
│   客户端请求      │    │    API接口层      │    │    领域层        │
│                 │    │                  │    │                 │
│ - 更新分数       │───▶│ - HTTP路由       │───▶│ - 跳表索引      │
│ - 查询排名       │    │ - 参数验证       │    │ - 前K名堆       │
│ - 获取榜单       │    │ - 响应格式化     │    │ - 缓存策略      │
└─────────────────┘    └──────────────────┘    └─────────┬───────┘
                                                         │
                                                         ▼
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   存储层         │    │    监控层         │    │    扩展层        │
│                 │    │                  │    │                 │
│ - 内存存储       │◀──│ - 性能指标       │───▶│ - 插件系统      │
│ - Redis缓存     │    │ - 业务监控       │    │ - 分片策略      │
│ - 持久化接口     │    │ - 告警系统       │    │ - 多算法支持    │
└─────────────────┘    └──────────────────┘    └─────────────────┘

