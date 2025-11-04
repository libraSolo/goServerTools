# GoTools

一个包含多模块的 Go 学习与示例仓库，涵盖排行榜、异步处理、发布订阅与定时任务等主题。各模块均可独立运行或在工作区内协同开发。

## 模块总览
- `chart/leaderboard`：轻量排行榜服务，基于跳表实现排名；提供 HTTP API（Gin），包含快照与 AOF 持久化示例。
- `chart/chart`：混合策略排行榜（跳表 + 前 K 最小堆 + 缓存），提供 TopN 高效读取与批量更新通道的实现。
- `chart/rank-system`：另一套排行榜实现与类型定义（供示例模块引用）。
- `async/*`、`pubsub/*`、`timer/*`：异步、发布订阅、定时任务相关的小型示例与工具。


## 目录参考
仓库包含以下目录（非完整列表）：
- `async/`：并发与异步示例
- `pubsub/`：通用发布订阅实现
- `timer/`：时间轮与定时任务
- `chart/`：排行榜相关实现与示例

摘自各种框架有意思 or 可以借鉴的工具总结

- 异步与主线程投递模块：请参见 [async.md](async/async.md)
- 发布订阅模块总结：请参见 [pubsub.md](pubsub/pubsub.md)
  - Trie 前缀树实现与总结：请参见 [trie.md](go-trie-tst/trie.md)
- 定时任务模块总结：请参见 [time.md](timer/timer.md)
  - crontab go精准定时任务：请参见 [trie.md](timer/crontab/crontab.md)
  - 时间轮定时任务：请参见 [trie.md](timer/timeWheel/timeWheel.md)

