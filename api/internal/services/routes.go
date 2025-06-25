package services

import (
	"github.com/Yaon-C2H8N2/bahclePlayer/internal/controllers"
	"github.com/Yaon-C2H8N2/bahclePlayer/internal/models"
	"github.com/gin-gonic/gin"
)

func MapRoutes(router *gin.Engine, pm *controllers.PlayersManager, aw *controllers.ApiWrapper, esp *controllers.EventSubPool, status *models.AppStatus) {
	router.POST("/login", func(c *gin.Context) {
		login(c, aw, esp)
	})
	router.GET("/logout", logout)

	router.GET("/player", func(c *gin.Context) {
		getPlayer(c, pm)
	})
	router.PUT("/player/currentPlaying", setCurrentPlaying)
	router.POST("/addVideo", func(c *gin.Context) {
		//TODO : change path to /player/addVideo (and maybe change the method to PUT)
		addVideos(c, pm)
	})

	//TODO : change path to /player/playlist
	router.GET("/playlist", getPlaylistAndQueue)
	//TODO : change path to /player/playlist
	router.DELETE("/playlist", deleteVideo)

	router.GET("/settings", func(c *gin.Context) {
		saveSettings(c, aw)
	})
	router.GET("/rewards", func(c *gin.Context) {
		getRewardsIds(c, aw)
	})

	router.GET("/overlays", getOverlays)
	router.GET("/overlays/events", getEventSocket)

	router.GET("/appinfo", func(c *gin.Context) {
		c.JSON(200, status)
	})
}
