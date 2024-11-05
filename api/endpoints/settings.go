package endpoints

import (
	"context"
	"github.com/Yaon-C2H8N2/bahclePlayer/controllers"
	"github.com/Yaon-C2H8N2/bahclePlayer/utils"
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
		"rewards": rewards,
	})
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

	conn := utils.GetConnection()
	defer conn.Close(context.Background())

	sql := `
		INSERT INTO users_config (user_id, config, value)
		VALUES ((SELECT user_id FROM users WHERE twitch_id = $1), 'PLAYLIST_REDEMPTION', $2),
       		   ((SELECT user_id FROM users WHERE twitch_id = $1), 'QUEUE_REDEMPTION', $3)
		ON CONFLICT (user_id, config) DO UPDATE
		SET value = excluded.value;
	`
	utils.DoRequest(conn, sql, userInfo.ID, playlistRedemption, queueRedemption)

	c.JSON(200, gin.H{
		"message": "Settings saved",
	})
}
