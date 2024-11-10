package models

import "time"

type Users struct {
	UserId         int       `json:"user_id"`
	Username       string    `json:"username"`
	TwitchId       string    `json:"twitch_id"`
	Token          string    `json:"token"`
	TokenCreatedAt time.Time `json:"token_created_at"`
}
