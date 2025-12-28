package models

type Leaderboard struct {
	ID         int `gorm:"primaryKey;column:id" json:"id"`
	UserID     int `gorm:"not null;column:user_id" json:"user_id"`
	TotalScore int `gorm:"not null;column:total_score" json:"total_score"`
	Rank       int `gorm:"column:rank" json:"rank"`

	User User `gorm:"foreignKey:UserID;references:ID" json:"user"`
}

func (Leaderboard) TableName() string {
	return "leaderboard"
}

type LeaderboardSlice []*Leaderboard
