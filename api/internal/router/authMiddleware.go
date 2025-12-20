package router

import (
	"github.com/Yaon-C2H8N2/bahclePlayer/internal/controllers"
	"github.com/Yaon-C2H8N2/bahclePlayer/internal/models"
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
