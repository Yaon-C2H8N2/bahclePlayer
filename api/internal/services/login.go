package services

import (
	"github.com/Yaon-C2H8N2/bahclePlayer/internal/controllers"
	"github.com/Yaon-C2H8N2/bahclePlayer/internal/models"
	"github.com/Yaon-C2H8N2/bahclePlayer/pkg/utils"
	"github.com/gin-gonic/gin"
)

var excludedPaths = map[string]bool{
	"/player":          true,
	"/login":           true,
	"/logout":          true,
	"/appinfo":         true,
	"/overlays/events": true,
}

func AuthMiddleware(c *gin.Context, aw *controllers.ApiWrapper) {
	if excludedPaths[c.Request.URL.Path] {
		c.Next()
		return
	}

	token := c.Request.Header.Get("Authorization")
	if token == "" || len(token) < 7 {
		c.JSON(401, gin.H{
			"error": "missing access_token",
		})
		c.Abort()
		return
	}
	token = token[7:]

	parsedToken, err := models.ValidateToken(token)
	if err != nil {
		c.JSON(401, gin.H{
			"error": err.Error(),
		})
		c.Abort()
		return
	}
	tokenClaims := parsedToken.Claims.(*models.JWTClaims)
	user, err := models.GetUserFromUserId(tokenClaims.UserId)
	if err != nil {
		c.JSON(401, gin.H{
			"error": err.Error(),
		})
		c.Abort()
		return
	}

	userInfo, err := aw.GetUserInfoFromToken(user.Token)
	if err != nil {
		c.JSON(401, gin.H{
			"error": err.Error(),
		})
		c.Abort()
		return
	}

	c.Set("User", user)
	c.Set("TwitchUser", userInfo)
	c.Next()
}

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
	err = rows.Err()
	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}
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

		userInfo, err := aw.GetUserInfoFromToken(userToken.AccessToken)
		if err != nil {
			c.JSON(500, gin.H{
				"error": err.Error(),
			})
			return
		}
		conn.Release()

		user, err = models.AddOrUpdateUser(models.Users{
			Username:     userInfo.DisplayName,
			TwitchId:     userInfo.ID,
			Token:        userToken.AccessToken,
			RefreshToken: userToken.RefreshToken,
		}, *userToken)

		if err != nil {
			c.JSON(500, gin.H{
				"error": err.Error(),
			})
			return
		}

		es, err := controllers.GetEventSub(aw, user.Token)

		if err != nil {
			c.JSON(500, gin.H{
				"error": err.Error(),
			})
			return
		}

		es.OnStarted(func() {
			es.DropAllSubscriptions(user.Token)
			es.InitSubscriptions(user.Token)
		})
		es.Start()
		eventSubs[user.TwitchId] = es
	}

	jwtToken, err := models.GenerateToken(user)
	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.Header("Set-Cookie", "token="+jwtToken+"; Path=/;")
	c.JSON(200, gin.H{
		"token": jwtToken,
		"user":  user,
	})
}

func logout(c *gin.Context) {
	c.Header("Set-Cookie", "token=; Path=/; Expires=Thu, 01 Jan 1970 00:00:00 GMT")
	c.JSON(200, gin.H{
		"message": "Logged out",
	})
}
