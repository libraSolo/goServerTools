package types

import "time"

const (
	// APIPrefix 是API路由的统一前缀。
	APIPrefix = "/api/v1"
	// DefaultPage 是分页查询中的默认页码。
	DefaultPage = 1
	// DefaultPageSize 是分页查询中的默认页面大小。
	DefaultPageSize = 20
	// MaxPageSize 是分页查询中允许的最大页面大小。
	MaxPageSize = 1000
)

const (
	// CacheTTLShort 是短时间缓存的过期时间（5分钟）。
	CacheTTLShort = 5 * time.Minute
	// CacheTTLMedium 是中等时间缓存的过期时间（30分钟）。
	CacheTTLMedium = 30 * time.Minute
	// CacheTTLLong 是长时间缓存的过期时间（2小时）。
	CacheTTLLong = 2 * time.Hour
)

const (
	// MaxLeaderboardSize 是排行榜允许的最大玩家数量。
	MaxLeaderboardSize = 1000000
	// MaxBatchUpdateSize 是批量更新分数的最大批次大小。
	MaxBatchUpdateSize = 1000
)

const (
	// MinPlayerID 是玩家ID的最小值。
	MinPlayerID = 1
	// MaxPlayerID 是玩家ID的最大值。
	MaxPlayerID = 1<<63 - 1
	// MinScore 是分数的最小值。
	MinScore = 0
	// MaxScore 是分数的最大值。
	MaxScore = 1000000000
)

const (
	// EnvConfigPath 是存储配置文件路径的环境变量键。
	EnvConfigPath = "RANK_CONFIG_PATH"
	// EnvLogLevel 是控制日志级别的环境变量键。
	EnvLogLevel = "RANK_LOG_LEVEL"
	// EnvServerPort 是配置服务器端口的环境变量键。
	EnvServerPort = "RANK_SERVER_PORT"
)

const (
	// HeaderTraceID 是用于分布式追踪的HTTP头。
	HeaderTraceID = "X-Trace-ID"
	// HeaderUserID 是用于传递用户ID的HTTP头。
	HeaderUserID = "X-User-ID"
	// HeaderContentType 是标准的Content-Type HTTP头。
	HeaderContentType = "Content-Type"
)

const (
	// CacheKeyTopRanks 是排行榜前N名缓存的键前缀。
	CacheKeyTopRanks = "rank:top:%s"
	// CacheKeyPlayerRank 是玩家排名缓存的键前缀。
	CacheKeyPlayerRank = "rank:player:%s:%d"
	// CacheKeyLeaderboard 是排行榜信息缓存的键前缀。
	CacheKeyLeaderboard = "rank:lb:%s"
	// CacheKeyNearbyRanks 是临近排名缓存的键前缀。
	CacheKeyNearbyRanks = "rank:nearby:%s:%d"
)

const (
	// CodeSuccess 表示操作成功的错误码。
	CodeSuccess = 0
	// CodeInvalidParams 表示参数无效的错误码。
	CodeInvalidParams = 10001
	// CodeNotFound 表示资源未找到的错误码。
	CodeNotFound = 10002
	// CodeInternalError 表示内部错误的错误码。
	CodeInternalError = 10003
	// CodeDuplicate 表示操作重复的错误码。
	CodeDuplicate = 10004
	// CodeUnauthorized 表示未经授权的错误码。
	CodeUnauthorized = 10005
)

// ErrorMessages 是错误码到错误消息的映射。
var ErrorMessages = map[int]string{
	CodeSuccess:       "成功",
	CodeInvalidParams: "参数错误",
	CodeNotFound:      "资源不存在",
	CodeInternalError: "内部错误",
	CodeDuplicate:     "重复操作",
	CodeUnauthorized:  "未授权",
}

// ContextKey 是用于在上下文中存储值的键类型。
type ContextKey string

const (
	// ContextKeyTraceID 是用于在上下文中存储追踪ID的键。
	ContextKeyTraceID ContextKey = "trace_id"
	// ContextKeyUserID 是用于在上下文中存储用户ID的键。
	ContextKeyUserID ContextKey = "user_id"
)

const (
	// MetricUpdateScoreDuration 是更新分数操作耗时的性能指标名称。
	MetricUpdateScoreDuration = "update_score_duration"
	// MetricGetRankDuration 是获取排名操作耗时的性能指标名称。
	MetricGetRankDuration = "get_rank_duration"
	// MetricLeaderboardSize 是排行榜大小的性能指标名称。
	MetricLeaderboardSize = "leaderboard_size"
	// MetricCacheHitRate 是缓存命中率的性能指标名称。
	MetricCacheHitRate = "cache_hit_rate"
)