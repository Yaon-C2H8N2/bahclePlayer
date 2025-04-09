package services

import (
	"github.com/Yaon-C2H8N2/bahclePlayer/internal/controllers"
	"github.com/Yaon-C2H8N2/bahclePlayer/internal/models"
	"github.com/Yaon-C2H8N2/bahclePlayer/internal/models/songRequests"
	"github.com/gin-gonic/gin"
	"regexp"
)

func addVideos(c *gin.Context, pm *controllers.PlayersManager, aw *controllers.ApiWrapper) {
	token := c.Request.Header.Get("Authorization")
	token = token[7:]
	if token == "" {
		c.JSON(400, gin.H{
			"error": "missing access_token",
		})
		return
	}

	videoAddRequest := models.VideoAddRequest{}
	err := c.BindJSON(&videoAddRequest)

	if err != nil {
		c.JSON(400, gin.H{
			"error": "Failed to bind login request",
		})
		return
	}

	userInfo, err := aw.GetUserInfoFromToken(token)
	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	re := regexp.MustCompile(`(?:youtu\.be/|youtube\.com/(?:embed/|v/|watch\?v=|watch\?.+&v=))([^&\n?#]+)`)
	match := re.FindStringSubmatch(videoAddRequest.Url)
	youtubeId := ""
	if len(match) > 1 {
		youtubeId = match[1]
	} else {
		c.JSON(400, gin.H{
			"error": "Invalid YouTube URL",
		})
		return
	}

	youtubeResults, err := controllers.GetVideoInfo(youtubeId)
	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}
	if len(youtubeResults.Items) == 0 {
		c.JSON(400, gin.H{
			"error": "No video found",
		})
		return
	}
	videoInfo := youtubeResults.Items[0]

	songRequest := songRequests.SongRequest{
		YoutubeID: videoInfo.Id,
		Title:     videoInfo.Snippet.Title,
		Duration:  videoInfo.ContentDetails.Duration,
		Thumbnail: videoInfo.Snippet.Thumbnails.Default.Url,
		Type:      videoAddRequest.Type,
	}

	newVideo, err := songRequests.InsertRequestInDatabase(songRequest, userInfo.ID)
	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	conn := pm.GetConnFromToken(token)

	if conn != nil {
		for _, cn := range conn {
			cn.WriteJSON(newVideo)
		}
	}

	c.JSON(200, gin.H{
		"message": "Video added",
	})
}
