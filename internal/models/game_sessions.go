package models

import "time"

type GameSession struct {
	ID        int       `gorm:"primaryKey;column:id" json:"id"`
	UserID    int       `gorm:"not null;column:user_id" json:"user_id"`
	Score     int       `gorm:"not null;column:score" json:"score"`
	GameMode  string    `gorm:"not null;column:game_mode" json:"game_mode"`
	Timestamp time.Time `gorm:"column:timestamp;autoCreateTime" json:"timestamp"`

	User User `gorm:"foreignKey:UserID;references:ID" json:"-"`
}

func (GameSession) TableName() string {
	return "game_sessions"
}
