package types

import "time"

// Response 是一个通用的API响应结构，用于统一返回格式。
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	TraceID string      `json:"trace_id,omitempty"` // 用于分布式追踪
}

// PageRequest 定义了分页请求的基础结构，用于API中的分页查询。
type PageRequest struct {
	Page     int `json:"page" form:"page" binding:"min=1"`
	PageSize int `json:"page_size" form:"page_size" binding:"min=1,max=1000"`
}

// PageResponse 定义了分页响应的结构，包含了分页查询的结果和元数据。
type PageResponse struct {
	Page      int         `json:"page"`
	PageSize  int         `json:"page_size"`
	Total     int64       `json:"total"`
	TotalPage int         `json:"total_page"`
	Data      interface{} `json:"data"`
}

// TimeRange 用于定义一个时间范围，常用于按时间过滤的查询。
type TimeRange struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

// LeaderboardType 是排行榜类型的枚举，定义了不同周期的排行榜。
type LeaderboardType int

const (
	LeaderboardTypeDaily LeaderboardType = iota + 1 // 每日排行榜
	LeaderboardTypeWeekly                           // 每周排行榜
	LeaderboardTypeMonthly                          // 每月排行榜
	LeaderboardTypeSeason                           // 赛季排行榜
)

// String 实现了 Stringer 接口，返回 LeaderboardType 的字符串表示。
func (t LeaderboardType) String() string {
	switch t {
	case LeaderboardTypeDaily:
		return "daily"
	case LeaderboardTypeWeekly:
		return "weekly"
	case LeaderboardTypeMonthly:
		return "monthly"
	case LeaderboardTypeSeason:
		return "season"
	default:
		return "unknown"
	}
}

// RewardTier 定义了排行榜中的奖励等级，根据排名范围给予不同奖励。
type RewardTier struct {
	MinRank int    `json:"min_rank"`
	MaxRank int    `json:tMaxRank"`
	Reward  string `json:"reward"`
}

// LeaderboardStatus 是排行榜状态的枚举。
type LeaderboardStatus string

const (
	StatusActive   LeaderboardStatus = "active"   // 活跃状态
	StatusInactive LeaderboardStatus = "inactive" // 非活跃状态
	StatusArchived LeaderboardStatus = "archived" // 已归档状态
)

// ScoreType 是分数类型的枚举，用于区分不同性质的分数。
type ScoreType int

const (
	ScoreTypeNormal  ScoreType = iota // 正常分数
	ScoreTypeBonus                    // 奖励分数
	ScoreTypePenalty                  // 惩罚分数
)

// Validate 验证分数是否在指定类型的有效范围内。
func (s ScoreType) Validate(score int64) bool {
	switch s {
	case ScoreTypeNormal:
		return score >= 0 && score <= 1000000
	case ScoreTypeBonus:
		return score >= 0 && score <= 50000
	case ScoreTypePenalty:
		return score >= -10000 && score <= 0
	default:
		return false
	}
}