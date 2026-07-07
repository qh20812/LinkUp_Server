package models

import "time"

type CallHidden struct {
	CallID    string    `json:"call_id" gorm:"primaryKey;column:call_id"`
	UserID    string    `json:"user_id" gorm:"primaryKey;column:user_id"`
	CreatedAt time.Time `json:"created_at"`
}
