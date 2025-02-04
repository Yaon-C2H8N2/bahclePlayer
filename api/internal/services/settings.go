package services

import (
	"context"
	"fmt"
	"github.com/Yaon-C2H8N2/bahclePlayer/internal/controllers"
	"github.com/Yaon-C2H8N2/bahclePlayer/internal/models/twitch"
	"github.com/Yaon-C2H8N2/bahclePlayer/pkg/utils"
	"github.com/gin-gonic/gin"
)

func getRewardsIds(c *gin.Context, aw *controllers.ApiWrapper) {
	token := c.Request.Header.Get("Authorization")
	token = token[7:]
	if token == "" {
		c.JSON(400, gin.H{
			"error": "missing access_token",
		})
		return
	}

	userInfo, err := aw.GetUserInfoFromToken(token)
	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	rewards, err := aw.GetChannelRewards(token, userInfo.ID)
	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"rewards":  rewards,
		"settings": getSettings(userInfo),
	})
}

func getSettings(user twitch.UserInfo) []struct{ Config, Value string } {
	conn := utils.GetConnection()
	defer conn.Close(context.Background())

	sql := `
			SELECT config, value
			FROM users_config
			JOIN users ON users_config.user_id = users.user_id
			WHERE users.twitch_id = $1
		`
	rows := utils.DoRequest(conn, sql, user.ID)
	var config []struct{ Config, Value string }
	for rows.Next() {
		var result struct{ Config, Value string }
		err := rows.Scan(&result.Config, &result.Value)
		if err != nil {
			fmt.Println("Failed to get user config")
			return nil
		}
		config = append(config, result)
	}

	return config
}

func saveSettings(c *gin.Context, aw *controllers.ApiWrapper) {
	token := c.Request.Header.Get("Authorization")
	token = token[7:]
	if token == "" {
		c.JSON(400, gin.H{
			"error": "missing access_token",
		})
		return
	}

	userInfo, err := aw.GetUserInfoFromToken(token)
	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	playlistRedemption := c.Query("playlist_redemption")
	queueRedemption := c.Query("queue_redemption")
	playlistMethod := c.Query("playlist_method")
	queueMethod := c.Query("queue_method")

	if playlistMethod == "" {
		playlistMethod = "POLL"
	}
	if queueMethod == "" {
		queueMethod = "POLL"
	}

	conn := utils.GetConnection()
	defer conn.Close(context.Background())

	sql := `
		INSERT INTO users_config (user_id, config, value)
		VALUES ((SELECT user_id FROM users WHERE twitch_id = $1), 'PLAYLIST_REDEMPTION', $2),
       		   ((SELECT user_id FROM users WHERE twitch_id = $1), 'QUEUE_REDEMPTION', $3),
       		   ((SELECT user_id FROM users WHERE twitch_id = $1), 'PLAYLIST_METHOD', $4),
       		   ((SELECT user_id FROM users WHERE twitch_id = $1), 'QUEUE_METHOD', $5)
		ON CONFLICT (user_id, config) DO UPDATE
		SET value = excluded.value;
	`
	utils.DoRequest(conn, sql, userInfo.ID, playlistRedemption, queueRedemption, playlistMethod, queueMethod)

	c.JSON(200, gin.H{
		"message": "Settings saved",
	})
}
