package controller

import (
	"fmt"
	"gaming-leaderboard/constants"
	"gaming-leaderboard/internal/controller/request"
	gameSessionsSvc "gaming-leaderboard/internal/game_sessions/service"
	leaderboardSvc "gaming-leaderboard/internal/leaderboard/service"
	"gaming-leaderboard/pkg/apperror"
	"gaming-leaderboard/pkg/response"

	"github.com/gin-gonic/gin"
)

type LeaderboardController struct {
	gameSessionsService *gameSessionsSvc.GameSessionsService
	leaderboardService  *leaderboardSvc.LeaderboardService
}

func NewLeaderboardController(
	gameSessionsService *gameSessionsSvc.GameSessionsService,
	leaderboardService *leaderboardSvc.LeaderboardService,
) *LeaderboardController {
	return &LeaderboardController{
		gameSessionsService: gameSessionsService,
		leaderboardService:  leaderboardService,
	}
}

func (c *LeaderboardController) CreateScore(ctx *gin.Context) {
	var req request.SubmitScoreRequest
	if err := ctx.ShouldBindBodyWithJSON(&req); err != nil {
		apperror.New(fmt.Errorf("invalid request body: %w", err), 400).AbortWithError(ctx)
		return
	}

	cusErr := c.gameSessionsService.CreateGameSession(ctx, req)
	if cusErr.Exists() {
		cusErr.AbortWithError(ctx)
		return
	}

	response.OK(ctx, nil)
	return
}

func (c *LeaderboardController) GetTopLeaderboard(ctx *gin.Context) {
	leaders, cusErr := c.leaderboardService.GetTopLeaderboards(ctx)
	if cusErr.Exists() {
		cusErr.AbortWithError(ctx)
		return
	}

	response.OK(ctx, leaders)
	return
}

func (c *LeaderboardController) GetUserRankByUserID(ctx *gin.Context) {
	rank, cusErr := c.leaderboardService.GetUserRankByUserID(ctx, ctx.Param(constants.UserID))
	if cusErr.Exists() {
		cusErr.AbortWithError(ctx)
		return
	}

	response.OK(ctx, rank)
	return
}
