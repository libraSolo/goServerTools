package api

import (
	"net/http"
	"rank-system/service"
	"rank-system/types"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Handler HTTP请求处理器
type Handler struct {
	rankService *service.RankService
}

// NewHandler 创建处理器
func NewHandler(rankService *service.RankService) *Handler {
	return &Handler{
		rankService: rankService,
	}
}

// CreateLeaderboard 创建排行榜
func (h *Handler) CreateLeaderboard(c *gin.Context) {
	var req types.CreateLeaderboardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.Response{
			Code:    types.CodeInvalidParams,
			Message: types.ErrorMessages[types.CodeInvalidParams],
		})
		return
	}

	if err := h.rankService.CreateLeaderboard(&req); err != nil {
		c.JSON(http.StatusInternalServerError, types.Response{
			Code:    types.CodeInternalError,
			Message: types.ErrorMessages[types.CodeInternalError],
		})
		return
	}

	c.JSON(http.StatusCreated, types.Response{
		Code:    types.CodeSuccess,
		Message: types.ErrorMessages[types.CodeSuccess],
	})
}

// UpdateScore 更新玩家分数
func (h *Handler) UpdateScore(c *gin.Context) {
	var req types.BatchUpdateScoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.Response{
			Code:    types.CodeInvalidParams,
			Message: types.ErrorMessages[types.CodeInvalidParams],
		})
		return
	}

	results, err := h.rankService.BatchUpdateScore(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.Response{
			Code:    types.CodeInternalError,
			Message: types.ErrorMessages[types.CodeInternalError],
		})
		return
	}

	c.JSON(http.StatusOK, types.Response{
		Code:    types.CodeSuccess,
		Message: types.ErrorMessages[types.CodeSuccess],
		Data:    results,
	})
}

// GetPlayerRank 获取玩家排名
func (h *Handler) GetPlayerRank(c *gin.Context) {
	leaderboardID := c.Query("leaderboard_id")
	playerIDStr := c.Query("player_id")

	if leaderboardID == "" || playerIDStr == "" {
		c.JSON(http.StatusBadRequest, types.Response{
			Code:    types.CodeInvalidParams,
			Message: types.ErrorMessages[types.CodeInvalidParams],
		})
		return
	}

	playerID, err := strconv.ParseInt(playerIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.Response{
			Code:    types.CodeInvalidParams,
			Message: types.ErrorMessages[types.CodeInvalidParams],
		})
		return
	}

	req := &types.QueryLeaderboardRequest{
		LeaderboardID: leaderboardID,
		PlayerID:      playerID,
	}

	response, err := h.rankService.GetPlayerRank(req)
	if err != nil {
		c.JSON(http.StatusNotFound, types.Response{
			Code:    types.CodeNotFound,
			Message: types.ErrorMessages[types.CodeNotFound],
		})
		return
	}

	c.JSON(http.StatusOK, types.Response{
		Code:    types.CodeSuccess,
		Message: types.ErrorMessages[types.CodeSuccess],
		Data:    response,
	})
}

// GetNearbyRanks 获取临近排名
func (h *Handler) GetNearbyRanks(c *gin.Context) {
	leaderboardID := c.Query("leaderboard_id")
	playerIDStr := c.Query("player_id")
	pageSizeStr := c.Query("page_size")

	if leaderboardID == "" || playerIDStr == "" {
		c.JSON(http.StatusBadRequest, types.Response{
			Code:    types.CodeInvalidParams,
			Message: types.ErrorMessages[types.CodeInvalidParams],
		})
		return
	}

	playerID, err := strconv.ParseInt(playerIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.Response{
			Code:    types.CodeInvalidParams,
			Message: types.ErrorMessages[types.CodeInvalidParams],
		})
		return
	}

	pageSize := types.DefaultPageSize
	if pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil {
			pageSize = ps
		}
	}

	req := &types.QueryLeaderboardRequest{
		LeaderboardID: leaderboardID,
		PlayerID:      playerID,
		PageSize:      pageSize,
	}

	response, err := h.rankService.GetNearbyRanks(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.Response{
			Code:    types.CodeInternalError,
			Message: types.ErrorMessages[types.CodeInternalError],
		})
		return
	}

	c.JSON(http.StatusOK, types.Response{
		Code:    types.CodeSuccess,
		Message: types.ErrorMessages[types.CodeSuccess],
		Data:    response,
	})
}

// GetTopRanks 获取前N名
func (h *Handler) GetTopRanks(c *gin.Context) {
	leaderboardID := c.Query("leaderboard_id")
	pageSizeStr := c.Query("page_size")

	if leaderboardID == "" {
		c.JSON(http.StatusBadRequest, types.Response{
			Code:    types.CodeInvalidParams,
			Message: types.ErrorMessages[types.CodeInvalidParams],
		})
		return
	}

	pageSize := types.DefaultPageSize
	if pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil {
			pageSize = ps
		}
	}

	req := &types.QueryLeaderboardRequest{
		LeaderboardID: leaderboardID,
		PageSize:      pageSize,
	}

	response, err := h.rankService.GetTopRanks(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.Response{
			Code:    types.CodeInternalError,
			Message: types.ErrorMessages[types.CodeInternalError],
		})
		return
	}

	c.JSON(http.StatusOK, types.Response{
		Code:    types.CodeSuccess,
		Message: types.ErrorMessages[types.CodeSuccess],
		Data:    response,
	})
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(router *gin.Engine) {
	api := router.Group(types.APIPrefix)
	{
		api.POST("/leaderboards", h.CreateLeaderboard)
		api.PUT("/scores", h.UpdateScore)
		api.GET("/player-rank", h.GetPlayerRank)
		api.GET("/nearby-ranks", h.GetNearbyRanks)
		api.GET("/top-ranks", h.GetTopRanks)
	}
}