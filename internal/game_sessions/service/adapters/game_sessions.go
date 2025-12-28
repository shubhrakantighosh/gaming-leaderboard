package adapters

import (
	"gaming-leaderboard/internal/controller/request"
	"gaming-leaderboard/internal/models"
)

func ConvertToGameSessionModel(req request.SubmitScoreRequest) *models.GameSession {
	return &models.GameSession{
		UserID:   req.UserID,
		Score:    req.Score,
		GameMode: req.GameMode,
	}
}
