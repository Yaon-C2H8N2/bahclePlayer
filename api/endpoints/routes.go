package endpoints

import (
	"github.com/Yaon-C2H8N2/bahclePlayer/controllers"
	"github.com/gin-gonic/gin"
)

func MapRoutes(router *gin.Engine) {
	pm := controllers.DefaultPlayersManager()

	router.GET("/login", login)
	router.GET("/player", func(c *gin.Context) {
		getPlayer(c, pm)
	})
}
