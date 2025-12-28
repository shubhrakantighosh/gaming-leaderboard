package service

import (
	"context"
	"log"
	"sync"
	"time"

	leaderboardRepo "gaming-leaderboard/internal/leaderboard/repository"
)

// LeaderboardWorker handles batch rank recalculation
type LeaderboardWorker struct {
	repository         *leaderboardRepo.LeaderboardRepository
	leaderboardService *LeaderboardService
	mu                 sync.Mutex
	batchInterval      time.Duration
}

// *optimise approach*//
// we add something any flag value so if anyone submit new score will mark this flag as true
// then it will trigger the recalculation in the next batch interval
// once done recalculation we will reset the flag to false
func NewLeaderboardWorker(
	repo *leaderboardRepo.LeaderboardRepository,
	leaderboardService *LeaderboardService,
	batchInterval time.Duration,
) *LeaderboardWorker {
	return &LeaderboardWorker{
		repository:         repo,
		leaderboardService: leaderboardService,
		batchInterval:      batchInterval,
	}
}

// Start begins the background worker
func (w *LeaderboardWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.batchInterval)

	go func() {
		defer ticker.Stop()

		log.Printf("[INFO] LeaderboardWorker started | batch_interval=%v", w.batchInterval)

		for {
			select {
			case <-ctx.Done():
				log.Println("[INFO] LeaderboardWorker context cancelled")
				return
			case <-ticker.C:
				w.processBatch(ctx)
			}
		}
	}()
}

// processBatch recalculates leaderboard if there are pending updates
func (w *LeaderboardWorker) processBatch(ctx context.Context) {
	w.mu.Lock()
	defer w.mu.Unlock()

	log.Printf("[INFO] Processing leaderboard recalculation")
	startTime := time.Now()

	// Recalculate all ranks
	if err := w.repository.RecalculateAllRanksWithIsolation(ctx); err != nil {
		log.Printf("[ERROR] Leaderboard recalculation failed | err=%v", err)
		return
	}

	duration := time.Since(startTime)
	log.Printf("[INFO] Leaderboard recalculation completed | duration=%v", duration)

	// Invalidate cache after successful recalculation
	if err := w.leaderboardService.InvalidateTopCache(ctx); err != nil {
		log.Printf("[WARN] Cache invalidation failed | err=%v", err)
	}
}
