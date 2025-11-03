package main

import (
	"log"
	"rank-system/api"
	"rank-system/domain"
	"rank-system/storage"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化存储
	repo := storage.NewMemoryRepository()

	// 创建默认排行榜
	config := &domain.RankConfig{
		TotalPlayers: 300000,
		RewardRatio:  0.003,
		MinReward:    100,
		MaxReward:    1000,
	}

	leaderboard := domain.NewLeaderboard("default", "默认排行榜", config)
	if err := repo.SaveLeaderboard(leaderboard); err != nil {
		log.Fatal("Failed to create default leaderboard:", err)
	}

	// 初始化处理器
	handler := api.NewHandler(repo)

	// 设置Gin
	router := gin.Default()

	// 注册路由
	handler.RegisterRoutes(router)

	// 启动服务
	log.Println("Server starting on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
