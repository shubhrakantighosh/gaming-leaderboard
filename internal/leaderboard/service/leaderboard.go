package service

import (
	"context"
	"fmt"
	"log"

	"gaming-leaderboard/constants"
	"gaming-leaderboard/internal/leaderboard/repository"
	"gaming-leaderboard/internal/models"
	"gaming-leaderboard/pkg/apperror"
	oredis "gaming-leaderboard/pkg/redis"

	"github.com/newrelic/go-agent/v3/newrelic"
	"gorm.io/gorm"
)

type LeaderboardService struct {
	repository  *repository.LeaderboardRepository
	redisClient oredis.Cache
}

func NewLeaderboardService(
	repository *repository.LeaderboardRepository,
	redisClient oredis.Cache,
) *LeaderboardService {
	return &LeaderboardService{
		repository:  repository,
		redisClient: redisClient,
	}
}

// GetTopLeaderboards retrieves top leaderboards with caching
func (s *LeaderboardService) GetTopLeaderboards(
	ctx context.Context,
) (models.LeaderboardSlice, apperror.Error) {

	txn := newrelic.FromContext(ctx)
	cacheKey := fmt.Sprintf(
		constants.LeaderboardTopKeyFormat,
		constants.TopLeaderboardLimit,
	)

	// Cache lookup
	var cachedLeaders models.LeaderboardSlice
	found, err := s.redisClient.Get(ctx, cacheKey, &cachedLeaders)
	if err == nil && found {
		log.Println("[INFO] leaderboard top fetched from cache")

		if txn != nil {
			txn.AddAttribute("cache_hit", true)
		}

		return cachedLeaders, apperror.Error{}
	}

	if err != nil {
		log.Printf("[WARN] leaderboard cache get failed | err=%v", err)
		if txn != nil {
			txn.NoticeError(err)
		}
	}

	leaders, cusErr := s.repository.GetAll(ctx, nil, func(db *gorm.DB) *gorm.DB {
		return db.Order("rank ASC").Limit(constants.TopLeaderboardLimit)
	})
	if cusErr.Exists() {
		if txn != nil {
			txn.NoticeError(cusErr)
		}
		return nil, cusErr
	}

	// Cache set (non-blocking)
	if _, err := s.redisClient.Set(ctx, cacheKey, leaders, constants.OneHour); err != nil {
		log.Printf("[WARN] leaderboard cache set failed | err=%v", err)
		if txn != nil {
			txn.NoticeError(err)
		}
	}

	return leaders, apperror.Error{}
}

// GetUserRankByUserID retrieves user rank with caching
func (s *LeaderboardService) GetUserRankByUserID(
	ctx context.Context,
	userID string,
) (models.Leaderboard, apperror.Error) {

	txn := newrelic.FromContext(ctx)
	cacheKey := fmt.Sprintf(
		constants.LeaderboardUserKeyFormat,
		userID,
	)

	// Cache lookup
	var cachedLeader models.Leaderboard
	found, err := s.redisClient.Get(ctx, cacheKey, &cachedLeader)
	if err == nil && found {
		if txn != nil {
			txn.AddAttribute("cache_hit", true)
			txn.AddAttribute("user_id", userID)
		}
		return cachedLeader, apperror.Error{}
	}

	if err != nil {
		log.Printf(
			"[WARN] user leaderboard cache get failed | user_id=%s | err=%v",
			userID,
			err,
		)
		if txn != nil {
			txn.NoticeError(err)
		}
	}

	// DB fetch
	filter := map[string]interface{}{
		constants.UserID: userID,
	}

	leader, cusErr := s.repository.Get(ctx, filter)
	if cusErr.Exists() {
		if txn != nil {
			txn.NoticeError(cusErr)
		}
		return models.Leaderboard{}, cusErr
	}

	// Cache set (non-blocking)
	if _, err := s.redisClient.Set(ctx, cacheKey, leader, constants.OneHour); err != nil {
		log.Printf(
			"[WARN] user leaderboard cache set failed | user_id=%s | err=%v",
			userID,
			err,
		)
		if txn != nil {
			txn.NoticeError(err)
		}
	}

	return leader, apperror.Error{}
}

// InvalidateUserCache
func (s *LeaderboardService) InvalidateUserCache(ctx context.Context, userID string) error {
	cacheKey := fmt.Sprintf(constants.LeaderboardUserKeyFormat, userID)

	if _, err := s.redisClient.Unlink(ctx, []string{cacheKey}); err != nil {
		if txn := newrelic.FromContext(ctx); txn != nil {
			txn.NoticeError(err)
		}
		return err
	}
	return nil
}

// InvalidateTopCache
func (s *LeaderboardService) InvalidateTopCache(ctx context.Context) error {
	cacheKey := fmt.Sprintf(
		constants.LeaderboardTopKeyFormat,
		constants.TopLeaderboardLimit,
	)

	if _, err := s.redisClient.Unlink(ctx, []string{cacheKey}); err != nil {
		if txn := newrelic.FromContext(ctx); txn != nil {
			txn.NoticeError(err)
		}
		return err
	}
	return nil
}
