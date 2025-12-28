package models

import "time"

type User struct {
	ID       int       `gorm:"primaryKey;column:id" json:"id"`
	Username string    `gorm:"unique;not null;column:username" json:"username"`
	JoinDate time.Time `gorm:"column:join_date;autoCreateTime" json:"join_date"`
}
