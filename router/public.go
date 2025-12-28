package router

import (
	"context"
	"gaming-leaderboard/internal/controller"
	gameSessionsRepo "gaming-leaderboard/internal/game_sessions/repository"
	gameSessionsSvc "gaming-leaderboard/internal/game_sessions/service"
	leaderboardRepo "gaming-leaderboard/internal/leaderboard/repository"
	leaderboardSvc "gaming-leaderboard/internal/leaderboard/service"
	"gaming-leaderboard/middleware"
	"gaming-leaderboard/pkg/db/postgres"
	"gaming-leaderboard/pkg/redis"
	"time"

	"github.com/gin-gonic/gin"
)

func RegisterPublicRoutes(ctx context.Context, engine *gin.Engine) {
	engine.GET("/health", gin.HandlerFunc(func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	}))

	// move to initialization
	leaderboardRepository := leaderboardRepo.NewLeaderboardRepository(postgres.GetCluster().DbCluster)
	gameSessionsRepository := gameSessionsRepo.NewGameSessionsRepository(postgres.GetCluster().DbCluster)

	leaderboardService := leaderboardSvc.NewLeaderboardService(leaderboardRepository, redis.GetClient())
	leaderboardWorker := leaderboardSvc.NewLeaderboardWorker(
		leaderboardRepository,
		leaderboardService,
		3*time.Minute,
	)

	leaderboardWorker.Start(ctx)

	gameSessionsService := gameSessionsSvc.NewGameSessionsService(
		gameSessionsRepository,
		leaderboardService,
		leaderboardWorker,
	)

	controller := controller.NewLeaderboardController(
		gameSessionsService,
		leaderboardService,
	)

	apiV1 := engine.Group("/api/v1/",
		middleware.CORSMiddleware(),
		middleware.NewRelicMiddleware(),
		middleware.SanitizeQueryParams(),
		middleware.RequestLogger())
	{

		leaderboard := apiV1.Group("/leaderboard")
		{
			leaderboard.POST("/submit", controller.CreateScore)
			leaderboard.GET("/top", controller.GetTopLeaderboard)
			leaderboard.GET("/rank/:user_id", controller.GetUserRankByUserID)
		}
	}
}
