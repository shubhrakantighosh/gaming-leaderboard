package request

type SubmitScoreRequest struct {
	UserID   int    `json:"user_id" binding:"required,gt=0"`
	Score    int    `json:"score" binding:"required,gt=0"`
	GameMode string `json:"game_mode" binding:"required,oneof=solo team"`
}
