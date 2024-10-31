package endpoints

import (
	"github.com/Yaon-C2H8N2/bahclePlayer/controllers"
	"github.com/gin-gonic/gin"
)

func getPlayer(c *gin.Context, pm *controllers.PlayersManager) {
	token := c.Query("access_token")
	if token == "" {
		c.JSON(400, gin.H{
			"error": "missing access_token",
		})
		return
	}

	pm.CreatePlayer(c)

	c.JSON(200, gin.H{
		"message": "Player created",
	})
}
