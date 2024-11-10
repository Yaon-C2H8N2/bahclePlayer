package endpoints

import (
	"context"
	"github.com/Yaon-C2H8N2/bahclePlayer/controllers"
	"github.com/Yaon-C2H8N2/bahclePlayer/models"
	"github.com/Yaon-C2H8N2/bahclePlayer/utils"
	"github.com/gin-gonic/gin"
)

func login(c *gin.Context, aw *controllers.ApiWrapper, eventSub *controllers.EventSub) {
	token := c.Query("access_token")

	if token == "" {
		c.JSON(400, gin.H{
			"error": "missing access_token",
		})
	}

	userInfo, err := aw.GetUserInfoFromToken(token)
	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	conn := utils.GetConnection()
	defer conn.Close(context.Background())

	sql := `
			SELECT * FROM users
			WHERE twitch_id = $1
		`
	rows := utils.DoRequest(conn, sql, userInfo.ID)
	var user models.Users
	if rows.Next() {
		rows.Scan(&user.TwitchId, &user.Username, &user.Token, &user.TokenCreatedAt)
		// TODO : refresh token if needed
	} else {
		sql = `
				INSERT INTO users (twitch_id, username, token, token_created_at)
				VALUES ($1, $2, $3, now())
				RETURNING *
			`
		rows = utils.DoRequest(conn, sql, userInfo.ID, userInfo.DisplayName, token)
		rows.Scan(&user.TwitchId, &user.Username, &user.Token, &user.TokenCreatedAt)

		eventSub.InitSubscriptions(token)
	}

	c.Header("Set-Cookie", "token="+token+"; Path=/;")
	c.JSON(200, gin.H{
		"token": token,
		"user":  user,
	})
}

func logout(c *gin.Context) {
	c.Header("Set-Cookie", "token=; Path=/; Expires=Thu, 01 Jan 1970 00:00:00 GMT")
	c.JSON(200, gin.H{
		"message": "Logged out",
	})
}
