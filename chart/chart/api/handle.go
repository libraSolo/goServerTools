package api

import (
    "net/http"
    "chart/storage"
    "strconv"

    "github.com/gin-gonic/gin"
)

// Handler HTTP请求处理器
type Handler struct {
	repo storage.Repository
}

// NewHandler 创建处理器
func NewHandler(repo storage.Repository) *Handler {
	return &Handler{
		repo: repo,
	}
}

// UpdateScore 更新玩家分数
func (h *Handler) UpdateScore(c *gin.Context) {
	leaderboardID := c.Query("leaderboard_id")
	if leaderboardID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "leaderboard_id is required"})
		return
	}

	var req struct {
		PlayerID int64 `json:"player_id" binding:"required"`
		Score    int64 `json:"score" binding:"required"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	leaderboard, err := h.repo.GetLeaderboard(leaderboardID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "leaderboard not found"})
		return
	}

	if err := leaderboard.UpdateScore(req.PlayerID, req.Score); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 保存更新后的排行榜
	if err := h.repo.SaveLeaderboard(leaderboard); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// GetPlayerRank 获取玩家排名
func (h *Handler) GetPlayerRank(c *gin.Context) {
	leaderboardID := c.Query("leaderboard_id")
	playerIDStr := c.Query("player_id")

	if leaderboardID == "" || playerIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "leaderboard_id and player_id are required"})
		return
	}

	playerID, err := strconv.ParseInt(playerIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid player_id"})
		return
	}

	leaderboard, err := h.repo.GetLeaderboard(leaderboardID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "leaderboard not found"})
		return
	}

	rank, err := leaderboard.GetPlayerRank(playerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"player_id": playerID,
		"rank":      rank,
	})
}

// GetTopRanks 获取前N名
func (h *Handler) GetTopRanks(c *gin.Context) {
	leaderboardID := c.Query("leaderboard_id")
	limitStr := c.DefaultQuery("limit", "100")

	if leaderboardID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "leaderboard_id is required"})
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 100
	}

	leaderboard, err := h.repo.GetLeaderboard(leaderboardID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "leaderboard not found"})
		return
	}

	topRanks := leaderboard.GetTopRanks(limit)
	c.JSON(http.StatusOK, topRanks)
}

// GetLeaderboardInfo 获取排行榜信息
func (h *Handler) GetLeaderboardInfo(c *gin.Context) {
	leaderboardID := c.Query("leaderboard_id")
	if leaderboardID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "leaderboard_id is required"})
		return
	}

	leaderboard, err := h.repo.GetLeaderboard(leaderboardID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "leaderboard not found"})
		return
	}

	playerCount := leaderboard.GetPlayerCount()

	c.JSON(http.StatusOK, gin.H{
		"id":           leaderboard.ID,
		"name":         leaderboard.Name,
		"player_count": playerCount,
		"config":       leaderboard.Config,
	})
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(router *gin.Engine) {
	api := router.Group("/api/v1")
	{
		api.PUT("/scores", h.UpdateScore)
		api.GET("/player-rank", h.GetPlayerRank)
		api.GET("/top-ranks", h.GetTopRanks)
		api.GET("/leaderboard", h.GetLeaderboardInfo)
	}
}
