package repository

import (
	"context"
	"database/sql"
	"log"

	"gaming-leaderboard/internal/models"
	"gaming-leaderboard/internal/repository"
	"gaming-leaderboard/pkg/db/postgres"
)

type LeaderboardRepository struct {
	repository.Interface[models.Leaderboard]
	db *postgres.DbCluster
}

func NewLeaderboardRepository(db *postgres.DbCluster) *LeaderboardRepository {
	return &LeaderboardRepository{
		Interface: &repository.Repository[models.Leaderboard]{Db: db},
		db:        db,
	}
}

// RecalculateAllRanksWithIsolation recalculates with proper concurrency handling
func (r *LeaderboardRepository) RecalculateAllRanksWithIsolation(ctx context.Context) error {
	tx := r.db.GetMasterDB(ctx).Begin(&sql.TxOptions{
		Isolation: sql.LevelRepeatableRead,
	})

	if tx.Error != nil {
		return tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	query := `
		WITH user_scores AS (
			SELECT 
				user_id,
				SUM(score) as total_score
			FROM game_sessions
			GROUP BY user_id
		),
		ranked_users AS (
			SELECT 
				user_id,
				total_score,
				RANK() OVER (ORDER BY total_score DESC) as new_rank
			FROM user_scores
		)
		INSERT INTO leaderboard (user_id, total_score, rank)
		SELECT user_id, total_score, new_rank FROM ranked_users
		ON CONFLICT (user_id)
		DO UPDATE SET
			total_score = EXCLUDED.total_score,
			rank = EXCLUDED.rank
	`

	if err := tx.Exec(query).Error; err != nil {
		tx.Rollback()
		log.Printf("[ERROR] RecalculateAllRanksWithIsolation: err=%v", err)
		return err
	}

	if err := tx.Commit().Error; err != nil {
		log.Printf("[ERROR] RecalculateAllRanksWithIsolation: commit failed | err=%v", err)
		return err
	}

	log.Printf("[INFO] RecalculateAllRanksWithIsolation: completed successfully")
	return nil
}
