package http

import (
    "leaderboard/internal/application"
    "net/http"
    "strconv"
    "time"

    "github.com/gin-gonic/gin"
)

// Handler 负责处理 HTTP 请求。
type Handler struct {
	rankService application.RankService
}

// NewHandler 创建一个新的 Handler。
func NewHandler(rankService application.RankService) *Handler {
	return &Handler{rankService: rankService}
}

// RegisterRoutes 注册路由。
func (h *Handler) RegisterRoutes(router *gin.Engine) {
	api := router.Group("/api/v1")
	{
		api.POST("/scores", h.updateScore)
		api.GET("/ranks/:playerID", h.getPlayerRank)
		api.GET("/ranks/top/:n", h.getTopN)
		api.GET("/ranks/nearby/:playerID/:count", h.getNearbyRanks)
	}
}

func (h *Handler) updateScore(c *gin.Context) {
	var req struct {
		PlayerID int64 `json:"player_id"`
		Score    int64 `json:"score"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.rankService.UpdateScore(req.PlayerID, req.Score); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

func (h *Handler) getPlayerRank(c *gin.Context) {
	playerID, err := strconv.ParseInt(c.Param("playerID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid player id"})
		return
	}

	rank, err := h.rankService.GetPlayerRank(playerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"rank": rank})
}

func (h *Handler) getTopN(c *gin.Context) {
    n, err := strconv.Atoi(c.Param("n"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid n"})
        return
    }

    players, err := h.rankService.GetTopN(n)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    type rankedPlayer struct {
        ID        int64     `json:"id"`
        Score     int64     `json:"score"`
        Rank      int64     `json:"rank"`
        UpdatedAt time.Time `json:"updated_at"`
    }

    resp := make([]rankedPlayer, 0, len(players))
    for _, p := range players {
        rank, err := h.rankService.GetPlayerRank(p.ID)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        resp = append(resp, rankedPlayer{
            ID:        p.ID,
            Score:     p.Score,
            Rank:      rank,
            UpdatedAt: p.UpdatedAt,
        })
    }

    c.JSON(http.StatusOK, resp)
}

func (h *Handler) getNearbyRanks(c *gin.Context) {
    playerID, err := strconv.ParseInt(c.Param("playerID"), 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid player id"})
        return
    }

	count, err := strconv.Atoi(c.Param("count"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid count"})
		return
	}

    players, err := h.rankService.GetNearbyRanks(playerID, count)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    type rankedPlayer struct {
        ID        int64     `json:"id"`
        Score     int64     `json:"score"`
        Rank      int64     `json:"rank"`
        UpdatedAt time.Time `json:"updated_at"`
    }
    resp := make([]rankedPlayer, 0, len(players))
    for _, p := range players {
        rank, err := h.rankService.GetPlayerRank(p.ID)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        resp = append(resp, rankedPlayer{
            ID:        p.ID,
            Score:     p.Score,
            Rank:      rank,
            UpdatedAt: p.UpdatedAt,
        })
    }

    c.JSON(http.StatusOK, resp)
}