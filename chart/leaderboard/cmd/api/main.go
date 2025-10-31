package main

import (
	"leaderboard/internal/application"
	"leaderboard/internal/infrastructure/persistence"
	"leaderboard/internal/interfaces/http"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	log.Println("Starting application...")
	// 初始化存储库
	lb, repo, err := persistence.NewLeaderboardRepository("./data", "default")
	if err != nil {
		log.Fatalf("failed to create repository: %v", err)
	}
	log.Println("Repository created.")

	// 初始化应用服务
	rankService, err := application.NewRankService(lb, repo)
	if err != nil {
		log.Fatalf("failed to create rank service: %v", err)
	}
	log.Println("Rank service created.")

	// 初始化 HTTP 处理器
	handler := http.NewHandler(rankService)
	log.Println("HTTP handler created.")

	// 初始化 Gin 引擎
	router := gin.Default()
	log.Println("Gin engine created.")

	// 注册路由
	handler.RegisterRoutes(router)
	log.Println("Routes registered.")

	// 启动服务器
	log.Println("Starting server on :8080...")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
	log.Println("Server stopped.")
}