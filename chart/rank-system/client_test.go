package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"rank-system/domain"
	"rank-system/service"
	"rank-system/storage"
	"sync"
	"testing"
	"time"
)

// TestClient 运行一个测试客户端，以验证排行榜系统的API功能。
func TestClient(t *testing.T) {
	// 测试创建排行榜
	testCreateLeaderboard()

	// 并发测试更新分数
	testConcurrentUpdates()

	// 测试查询功能
	testQueries()
}

// testCreateLeaderboard 测试创建排行榜的功能。
func testCreateLeaderboard() {
	req := map[string]interface{}{
		"id":            "test",
		"name":          "测试排行榜",
		"total_players": 100000,
		"reward_ratio":  0.01,
		"min_reward":    100,
		"max_reward":    1000,
	}

	body, _ := json.Marshal(req)
	resp, err := http.Post("http://localhost:8080/api/v1/leaderboards", "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("创建排行榜失败: %v", err)
		return
	}
	defer resp.Body.Close()

	log.Printf("创建排行榜响应: %s", resp.Status)
}

// testConcurrentUpdates 测试并发更新分数的功能。
func testConcurrentUpdates() {
	const numPlayers = 1000
	const numUpdates = 10

	var wg sync.WaitGroup
	start := time.Now()

	for i := 0; i < numUpdates; i++ {
		wg.Add(1)
		go func(batch int) {
			defer wg.Done()

			for j := 0; j < numPlayers; j++ {
				playerID := int64(batch*numPlayers + j)
				score := int64((playerID % 500) + 1000)

				updateScore("default", playerID, score)
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	log.Printf("并发更新 %d 个玩家分数完成，耗时: %v", numPlayers*numUpdates, duration)
}

// updateScore 更新指定玩家的分数。
func updateScore(leaderboardID string, playerID, score int64) {
	req := map[string]interface{}{
		"player_id": playerID,
		"score":     score,
	}

	body, _ := json.Marshal(req)
	url := fmt.Sprintf("http://localhost:8080/api/v1/scores?leaderboard_id=%s", leaderboardID)

	request, _ := http.NewRequest("PUT", url, bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(request)
	if err != nil {
		log.Printf("更新分数失败: %v", err)
		return
	}
	defer resp.Body.Close()
}

// testQueries 测试所有查询功能。
func testQueries() {
	// 测试获取玩家排名
	testGetPlayerRank("default", 123)

	// 测试获取临近排名
	testGetNearbyRanks("default", 123, 3)

	// 测试获取前N名
	testGetTopRanks("default")
}

// testGetPlayerRank 测试获取单个玩家排名的功能。
func testGetPlayerRank(leaderboardID string, playerID int64) {
	url := fmt.Sprintf("http://localhost:8080/api/v1/player-rank?leaderboard_id=%s&player_id=%d", leaderboardID, playerID)

	resp, err := http.Get(url)
	if err != nil {
		log.Printf("获取玩家排名失败: %v", err)
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	log.Printf("玩家排名响应: %v", result)
}

// testGetNearbyRanks 测试获取临近排名的功能。
func testGetNearbyRanks(leaderboardID string, playerID int64, rangeSize int) {
	url := fmt.Sprintf("http://localhost:8080/api/v1/nearby-ranks?leaderboard_id=%s&player_id=%d&range_size=%d",
		leaderboardID, playerID, rangeSize)

	resp, err := http.Get(url)
	if err != nil {
		log.Printf("获取临近排名失败: %v", err)
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("解析临近排名响应失败: %v", err)
		return
	}

	if data, ok := result["data"].([]interface{}); ok {
		log.Printf("临近排名响应，数据条数: %d", len(data))
	} else {
		log.Printf("临近排名响应不包含有效数据: %v", result)
	}
}

// testGetTopRanks 测试获取排行榜前N名的功能。
func testGetTopRanks(leaderboardID string) {
	url := fmt.Sprintf("http://localhost:8080/api/v1/top-ranks?leaderboard_id=%s", leaderboardID)

	resp, err := http.Get(url)
	if err != nil {
		log.Printf("获取前N名失败: %v", err)
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("解析前N名响应失败: %v", err)
		return
	}

	if data, ok := result["data"].([]interface{}); ok {
		log.Printf("前N名响应，数据条数: %d", len(data))
	} else {
		log.Printf("前N名响应不包含有效数据: %v", result)
	}
}

// BenchmarkRankSystem 对排行榜系统的核心功能进行基准测试。
func BenchmarkRankSystem(b *testing.B) {
	repo := storage.NewMemoryRepository()
	rankService := service.NewRankService(repo)

	// 创建测试数据
	leaderboard := domain.NewLeaderboard("benchmark", "Benchmark",
		domain.NewRankConfig(100000, 0.01, 100, 1000))

	for i := 0; i < 100000; i++ {
		leaderboard.UpdatePlayerScore(int64(i), int64(i%500+1000))
	}

	repo.Save(leaderboard)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		playerID := int64(i % 100000)
		score := int64((i % 600) + 1000)

		req := &service.UpdateScoreRequest{
			LeaderboardID: "benchmark",
			PlayerID:      playerID,
			Score:         score,
		}

		rankService.UpdateScore(req)
	}
}