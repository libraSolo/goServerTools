```
rank-system/
├── domain/           # 领域层
│   ├── leaderboard.go
│   └── player.go
├── service/          # 应用服务层
│   └── rank_service.go
├── storage/          # 基础设施层
│   ├── repository.go
│   └── memory.go
├── api/              # 接口层
│   └── handlers.go
├── types/            # 共享类型
│   └── types.go
└── main.go           # 程序入口
```