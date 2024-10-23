package endpoints

import (
	"github.com/Yaon-C2H8N2/bahclePlayer/models"
	"github.com/gin-gonic/gin"
)

func MapRoutes(router *gin.Engine) {
	pm := models.DefaultPlayersManager()

	router.GET("/login", login)
	router.GET("/player", func(c *gin.Context) {
		getPlayer(c, pm)
	})
}
