package constants

import "time"

const (
	RequestID                = "request_id"
	Env                      = "env"
	Consistency              = "consistency"
	EventualConsistency      = "eventual"
	StrongConsistency        = "strong"
	UserID                   = "user_id"
	OneMinute                = time.Minute
	OneHour                  = OneMinute * 60
	OneDay                   = OneHour * 24
	TopLeaderboardLimit      = 10
	LeaderboardTopKeyFormat  = "leaderboard:top:%d"
	LeaderboardUserKeyFormat = "leaderboard:user:%s"
)
