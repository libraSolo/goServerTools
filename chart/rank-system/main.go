package main

import (
	"fmt"
	"log"
	"rank-system/api"
	"rank-system/service"
	"rank-system/storage"
	"rank-system/types"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化依赖
	repo := storage.NewMemoryRepository()
	rankService := service.NewRankService(repo)
	handler := api.NewHandler(rankService)

	// 创建默认排行榜
	createDefaultLeaderboard(rankService)

	// 设置Gin路由
	router := gin.Default()

	// 注册路由
	handler.RegisterRoutes(router)

	// 启动服务
	addr := fmt.Sprintf(":%d", types.DefaultServerPort)
	log.Printf("Server starting on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

// createDefaultLeaderboard 创建默认排行榜
func createDefaultLeaderboard(rankService *service.RankService) {
	req := &types.CreateLeaderboardRequest{
		ID:           "default",
		Name:         "默认排行榜",
		TotalPlayers: 300000,
		RewardRatio:  0.003, // 0.3%
		MinReward:    100,
		MaxReward:    1000,
	}

	if err := rankService.CreateLeaderboard(req); err != nil {
		log.Printf("Failed to create default leaderboard: %v", err)
	} else {
		log.Println("Default leaderboard created successfully")
	}
}