package repository

import (
	"gaming-leaderboard/internal/models"
	"gaming-leaderboard/internal/repository"
	"gaming-leaderboard/pkg/db/postgres"
)

type GameSessionsRepository struct {
	repository.Interface[models.GameSession]
	db *postgres.DbCluster
}

func NewGameSessionsRepository(db *postgres.DbCluster) *GameSessionsRepository {
	return &GameSessionsRepository{
		Interface: &repository.Repository[models.GameSession]{Db: db},
		db:        db,
	}
}
