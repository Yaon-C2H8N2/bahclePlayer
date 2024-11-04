package endpoints

import (
	"github.com/Yaon-C2H8N2/bahclePlayer/controllers"
	"github.com/gin-gonic/gin"
	"os"
)

func MapRoutes(router *gin.Engine) {
	pm := controllers.DefaultPlayersManager()
	aw := controllers.GetApiWrapper()
	aw.SetClientId(os.Getenv("TWITCH_CLIENT_ID"))

	router.GET("/login", func(c *gin.Context) {
		login(c, aw)
	})
	router.GET("/logout", func(c *gin.Context) {
		logout(c)
	})
	router.GET("/player", func(c *gin.Context) {
		getPlayer(c, pm)
	})
	router.GET("/playlist", func(c *gin.Context) {
		getPlaylistAndQueue(c, aw)
	})
	router.DELETE("/playlist", func(c *gin.Context) {
		deleteVideo(c, aw)
	})
	router.GET("/settings", func(c *gin.Context) {
		saveSettings(c, aw)
	})
	router.GET("/rewards", func(c *gin.Context) {
		getRewardsIds(c, aw)
	})
}
