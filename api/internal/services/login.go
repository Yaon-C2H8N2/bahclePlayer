package services

import (
	"github.com/Yaon-C2H8N2/bahclePlayer/internal/controllers"
	"github.com/Yaon-C2H8N2/bahclePlayer/internal/models"
	"github.com/Yaon-C2H8N2/bahclePlayer/pkg/utils"
	"github.com/gin-gonic/gin"
)

func login(c *gin.Context, aw *controllers.ApiWrapper, eventSubs map[string]*controllers.EventSub) {
	loginRequest := models.LoginRequest{}
	err := c.BindJSON(&loginRequest)

	if err != nil {
		c.JSON(400, gin.H{
			"error": "Failed to bind login request",
		})
		return
	}

	conn := utils.GetConnection()
	sql := `
			INSERT INTO token_requests (code, requested_at)
			VALUES ($1, now())
			ON CONFLICT (code) DO NOTHING
			RETURNING code
		`
	rows := utils.DoRequest(conn, sql, loginRequest.Code)
	var user models.Users
	if !rows.Next() {
		c.JSON(401, gin.H{
			"error": "A token request with this code already exists",
		})
		return
	} else {
		userToken, err := controllers.RequestUserToken(loginRequest.Code)
		if err != nil {
			c.JSON(500, gin.H{
				"error": err.Error(),
			})
			return
		}

		conn.Release()
		conn := utils.GetConnection()
		sql = `
			DELETE FROM token_requests
			WHERE code = $1
		`
		utils.DoRequest(conn, sql, loginRequest.Code)

		userInfo, err := aw.GetUserInfoFromToken(userToken)
		if err != nil {
			c.JSON(500, gin.H{
				"error": err.Error(),
			})
			return
		}
		conn.Release()
		conn = utils.GetConnection()
		defer conn.Release()
		sql = `
				INSERT INTO users (twitch_id, username, token, token_created_at)
				VALUES ($1, $2, $3, now())
				ON CONFLICT (twitch_id) DO UPDATE SET token = $3, token_created_at = now()
				RETURNING twitch_id, username, token, token_created_at
			`
		rows = utils.DoRequest(conn, sql, userInfo.ID, userInfo.DisplayName, userToken)
		rows.Next()
		err = rows.Scan(&user.TwitchId, &user.Username, &user.Token, &user.TokenCreatedAt)

		if err != nil {
			c.JSON(500, gin.H{
				"error": err.Error(),
			})
			return
		}

		es := controllers.GetEventSub(aw, user.Token)
		es.OnStarted(func() {
			es.DropAllSubscriptions(user.Token)
			es.InitSubscriptions(user.Token)
		})
		es.Start()
		eventSubs[user.Token] = es
	}

	c.Header("Set-Cookie", "token="+user.Token+"; Path=/;")
	c.JSON(200, gin.H{
		"token": user.Token,
		"user":  user,
	})
}

func logout(c *gin.Context) {
	c.Header("Set-Cookie", "token=; Path=/; Expires=Thu, 01 Jan 1970 00:00:00 GMT")
	c.JSON(200, gin.H{
		"message": "Logged out",
	})
}
