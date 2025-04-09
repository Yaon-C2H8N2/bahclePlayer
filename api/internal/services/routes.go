package services

import (
	"github.com/Yaon-C2H8N2/bahclePlayer/internal/controllers"
	"github.com/Yaon-C2H8N2/bahclePlayer/internal/models"
	"github.com/gin-gonic/gin"
)

func MapRoutes(router *gin.Engine, pm *controllers.PlayersManager, aw *controllers.ApiWrapper, es map[string]*controllers.EventSub, status *models.AppStatus) {
	router.POST("/login", func(c *gin.Context) {
		login(c, aw, es)
	})
	router.GET("/logout", func(c *gin.Context) {
		logout(c)
	})
	router.GET("/player", func(c *gin.Context) {
		getPlayer(c, pm)
	})
	router.GET("/playlist", func(c *gin.Context) {
		getPlaylistAndQueue(c)
	})
	router.DELETE("/playlist", func(c *gin.Context) {
		deleteVideo(c)
	})
	router.GET("/settings", func(c *gin.Context) {
		saveSettings(c, aw)
	})
	router.GET("/rewards", func(c *gin.Context) {
		getRewardsIds(c, aw)
	})
	router.POST("/addVideo", func(c *gin.Context) {
		addVideos(c, pm, aw)
	})
	router.GET("/appinfo", func(c *gin.Context) {
		c.JSON(200, status)
	})
}
