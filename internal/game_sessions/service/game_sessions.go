package service

import (
	"context"
	"fmt"
	"strconv"

	"gaming-leaderboard/internal/controller/request"
	"gaming-leaderboard/internal/game_sessions/repository"
	"gaming-leaderboard/internal/game_sessions/service/adapters"
	leaderboardSvc "gaming-leaderboard/internal/leaderboard/service"
	"gaming-leaderboard/pkg/apperror"

	"github.com/newrelic/go-agent/v3/newrelic"
)

type GameSessionsService struct {
	repository         *repository.GameSessionsRepository
	leaderboardService *leaderboardSvc.LeaderboardService
	leaderboardWorker  *leaderboardSvc.LeaderboardWorker
}

func NewGameSessionsService(
	repo *repository.GameSessionsRepository,
	leaderboardService *leaderboardSvc.LeaderboardService,
	leaderboardWorker *leaderboardSvc.LeaderboardWorker,
) *GameSessionsService {
	return &GameSessionsService{
		repository:         repo,
		leaderboardService: leaderboardService,
		leaderboardWorker:  leaderboardWorker,
	}
}

func (s *GameSessionsService) CreateGameSession(
	ctx context.Context,
	sessionData request.SubmitScoreRequest,
) apperror.Error {
	txn := newrelic.FromContext(ctx)

	if cusErr := s.repository.Create(
		ctx,
		adapters.ConvertToGameSessionModel(sessionData),
	); cusErr.Exists() {
		if txn != nil {
			txn.NoticeError(cusErr)
		}
		return apperror.New(
			fmt.Errorf("unable to create session, please try again later"),
			400,
		)
	}

	s.leaderboardService.InvalidateUserCache(ctx, strconv.Itoa(sessionData.UserID))

	return apperror.Error{}
}
