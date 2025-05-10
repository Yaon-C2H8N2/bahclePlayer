package models

import (
	"fmt"
	"github.com/Yaon-C2H8N2/bahclePlayer/internal/models/twitch"
	"github.com/Yaon-C2H8N2/bahclePlayer/pkg/utils"
	"time"
)

type Users struct {
	UserId         int       `json:"user_id"`
	Username       string    `json:"username"`
	TwitchId       string    `json:"twitch_id"`
	Token          string    `json:"token"`
	TokenCreatedAt time.Time `json:"token_created_at"`
	TokenExpiresAt time.Time `json:"token_expires_at"`
	RefreshToken   string    `json:"refresh_token"`
}

func GetAllUsers() ([]Users, error) {
	conn := utils.GetConnection()
	defer conn.Release()
	sql := `
			SELECT users.user_id, users.twitch_id, users.username, users.token, users.token_created_at, users.token_expires_at, users.refresh_token
			FROM users
		`
	rows := utils.DoRequest(conn, sql)
	var users []Users
	for rows.Next() {
		var result Users
		err := rows.Scan(&result.UserId, &result.TwitchId, &result.Username, &result.Token, &result.TokenCreatedAt, &result.TokenExpiresAt, &result.RefreshToken)
		if err != nil {
			return []Users{}, err
		}
		users = append(users, result)
	}
	return users, nil
}

func GetUserFromUserId(userid int) (Users, error) {
	conn := utils.GetConnection()
	defer conn.Release()
	sql := `
			SELECT users.user_id, users.twitch_id, users.username, users.token, users.token_created_at, users.token_expires_at, users.refresh_token
			FROM users
			WHERE user_id = $1
		`
	rows := utils.DoRequest(conn, sql, userid)
	var user Users
	if !rows.Next() {
		return Users{}, nil
	}

	err := rows.Scan(&user.UserId, &user.TwitchId, &user.Username, &user.Token, &user.TokenCreatedAt, &user.TokenExpiresAt, &user.RefreshToken)
	if err != nil {
		return Users{}, err
	}
	return user, nil
}

func GetUserFromToken(token string) (Users, error) {
	var user Users
	conn := utils.GetConnection()
	defer conn.Release()
	sql := `
			SELECT users.user_id, users.twitch_id, users.username, users.token, users.token_created_at, users.token_expires_at, users.refresh_token
			FROM users
			WHERE token = $1
		`
	rows := utils.DoRequest(conn, sql, token)
	if !rows.Next() {
		return Users{}, fmt.Errorf("No user found with this token")
	}
	err := rows.Scan(&user.UserId, &user.TwitchId, &user.Username, &user.Token, &user.TokenCreatedAt, &user.TokenExpiresAt, &user.RefreshToken)
	if err != nil {
		return Users{}, err
	}
	return user, nil
}

func AddOrUpdateUser(user Users, userToken twitch.UserTokenResponse) (Users, error) {
	var resultUser Users
	conn := utils.GetConnection()
	defer conn.Release()
	sql := `
				INSERT INTO users (twitch_id, username, token, token_created_at, token_expires_at, refresh_token)
				VALUES ($1, $2, $3, $4, $5, $6)
				ON CONFLICT (twitch_id) DO UPDATE SET token = $3, token_created_at = $4, token_expires_at = $5, refresh_token = $6
				RETURNING user_id, twitch_id, username, token, token_created_at, token_expires_at, refresh_token
			`
	rows := utils.DoRequest(conn, sql, user.TwitchId, user.Username, userToken.AccessToken, time.Now(), time.Now().Add(time.Duration(userToken.ExpiresIn)*time.Second), userToken.RefreshToken)
	rows.Next()
	err := rows.Scan(&resultUser.UserId, &resultUser.TwitchId, &resultUser.Username, &resultUser.Token, &resultUser.TokenCreatedAt, &resultUser.TokenExpiresAt, &resultUser.RefreshToken)

	if err != nil {
		return Users{}, err
	}
	return resultUser, nil
}
