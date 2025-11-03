package domain

import (
	"testing"
	"time"
)

// helper: 创建并预置一个排行榜，含同分数的先后顺序判断
func setupLeaderboardBasic() *HybridLeaderboard {
	lb := NewHybridLeaderboard("test", "测试榜", &RankConfig{TotalPlayers: 1000})

	// 使用同步更新，避免批处理通道带来的异步性
	// 插入两个同分玩家，保证先插入者排序更靠前
	_ = lb.syncUpdateScore(2, 50)
	time.Sleep(1 * time.Millisecond)
	_ = lb.syncUpdateScore(4, 50)
	_ = lb.syncUpdateScore(3, 20)
	_ = lb.syncUpdateScore(1, 10)
	_ = lb.syncUpdateScore(5, 5)

	return lb
}

func idsOf(players []*Player) []int64 {
	ids := make([]int64, 0, len(players))
	for _, p := range players {
		ids = append(ids, p.ID)
	}
	return ids
}

func containsAll(ids []int64, expect []int64) bool {
	set := make(map[int64]bool, len(ids))
	for _, id := range ids {
		set[id] = true
	}
	for _, e := range expect {
		if !set[e] {
			return false
		}
	}
	return true
}

// 基础排名校验：分数高者排前，同分先更新者排前
func TestLeaderboardRankingBasic(t *testing.T) {
	lb := setupLeaderboardBasic()

	cases := []struct {
		id   int64
		want int
	}{
		{2, 1}, // 50 分，先更新者
		{4, 2}, // 50 分，后更新者
		{3, 3}, // 20 分
		{1, 4}, // 10 分
		{5, 5}, //  5 分
	}

	for _, c := range cases {
		got, err := lb.GetPlayerRank(c.id)
		if err != nil {
			t.Fatalf("GetPlayerRank(%d) error: %v", c.id, err)
		}
		if got != c.want {
			t.Fatalf("rank mismatch for %d: got=%d want=%d", c.id, got, c.want)
		}
	}
}

// TopN 校验：集合包含期望 ID（不强依赖切片内顺序）
func TestLeaderboardTopRanks(t *testing.T) {
	lb := setupLeaderboardBasic()
	top := lb.GetTopRanks(3)
	if len(top) != 3 {
		t.Fatalf("TopRanks length mismatch: got=%d want=3", len(top))
	}
	ids := idsOf(top)
	if !containsAll(ids, []int64{2, 4, 3}) {
		t.Fatalf("TopRanks IDs mismatch: got=%v expect contains {2,4,3}", ids)
	}
}

// 临近排名：以玩家 3 为中心，rangeSize=1，返回 [rank-1, rank, rank+1]
func TestLeaderboardNearbyRanks(t *testing.T) {
	lb := setupLeaderboardBasic()
	near, err := lb.GetNearbyRanks(3, 1)
	if err != nil {
		t.Fatalf("GetNearbyRanks error: %v", err)
	}
	if len(near) != 3 {
		t.Fatalf("Nearby length mismatch: got=%d want=3", len(near))
	}
	// 期望顺序：rank2(id=4), rank3(id=3), rank4(id=1)
	if near[0].ID != 4 || near[1].ID != 3 || near[2].ID != 1 {
		t.Fatalf("Nearby IDs order mismatch: got=%v want=[4,3,1]", idsOf(near))
	}
}

// 分数更新后排名应调整；同时验证 TopN 集合包含更新后的玩家
func TestLeaderboardUpdateScoreAdjustRank(t *testing.T) {
	lb := setupLeaderboardBasic()
	// 玩家 1 从 10 -> 60，应该升至第一名
	if err := lb.syncUpdateScore(1, 60); err != nil {
		t.Fatalf("syncUpdateScore error: %v", err)
	}
	r, err := lb.GetPlayerRank(1)
	if err != nil {
		t.Fatalf("GetPlayerRank error: %v", err)
	}
	if r != 1 {
		t.Fatalf("rank of player 1 mismatch: got=%d want=1", r)
	}

	top := lb.GetTopRanks(3)
	ids := idsOf(top)
	if !containsAll(ids, []int64{1, 2, 4}) { // 顶部应包含 1, 2, 4 三人
		t.Fatalf("TopRanks after update mismatch: got=%v expect contains {1,2,4}", ids)
	}
}

// 未找到玩家时应返回错误
func TestLeaderboardRankNotFound(t *testing.T) {
	lb := setupLeaderboardBasic()
	if _, err := lb.GetPlayerRank(99999); err == nil {
		t.Fatalf("expected error when player not found, got nil")
	}
}

// 缓存失效验证：获取 TopN 后更新分数应失效缓存，随后读取包含新高分玩家
func TestLeaderboardCacheInvalidateOnUpdate(t *testing.T) {
	lb := setupLeaderboardBasic()
	_ = lb.GetTopRanks(3) // 填充缓存
	if err := lb.syncUpdateScore(5, 100); err != nil {
		t.Fatalf("syncUpdateScore error: %v", err)
	}
	top := lb.GetTopRanks(3)
	ids := idsOf(top)
	if !containsAll(ids, []int64{5}) {
		t.Fatalf("TopRanks should contain player 5 after update, got=%v", ids)
	}
}
