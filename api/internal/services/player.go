package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Yaon-C2H8N2/bahclePlayer/internal/controllers"
	"github.com/Yaon-C2H8N2/bahclePlayer/internal/models"
	"github.com/Yaon-C2H8N2/bahclePlayer/internal/models/twitch"
	"github.com/Yaon-C2H8N2/bahclePlayer/pkg/utils"
	"github.com/gin-gonic/gin"
	"regexp"
)

func getPlayer(c *gin.Context, pm *controllers.PlayersManager) {
	pm.CreatePlayer(c)
}

func getPlaylistAndQueue(c *gin.Context) {
	TwitchUserContext, _ := c.Get("TwitchUser")
	userInfo, _ := TwitchUserContext.(twitch.UserInfo)

	conn := utils.GetConnection()
	defer conn.Release()

	sql := `
			SELECT users_videos.video_id, users_videos.user_id, users_videos.youtube_id, users_videos.url, users_videos.title, users_videos.duration, users_videos.type, users_videos.created_at, users_videos.thumbnail_url, users_videos.added_by
			FROM users_videos
			JOIN users ON users_videos.user_id = users.user_id
			WHERE users.twitch_id = $1
			ORDER BY users_videos.created_at DESC
		`
	rows := utils.DoRequest(conn, sql, userInfo.ID)
	var results []models.UsersVideos
	var err error
	for rows.Next() {
		var result models.UsersVideos
		err = rows.Scan(&result.VideoId, &result.UserId, &result.YoutubeId, &result.Url, &result.Title, &result.Duration, &result.Type, &result.CreatedAt, &result.ThumbnailUrl, &result.AddedBy)
		if err != nil {
			c.JSON(500, gin.H{
				"error": err.Error(),
			})
			return
		}
		results = append(results, result)
	}

	c.JSON(200, gin.H{
		"message": "Playlist and queue retrieved",
		"data":    results,
	})
}

func deleteVideo(c *gin.Context) {
	TwitchUserContext, _ := c.Get("TwitchUser")
	userInfo, _ := TwitchUserContext.(twitch.UserInfo)

	videoID := c.Query("video_id")
	if videoID == "" {
		c.JSON(400, gin.H{
			"error": "missing video_id",
		})
		return
	}

	conn := utils.GetConnection()
	defer conn.Release()

	sql := `
			DELETE FROM users_videos
			WHERE user_id = (SELECT user_id FROM users WHERE twitch_id = $1) AND video_id = $2
		`
	utils.DoRequest(conn, sql, userInfo.ID, videoID)

	c.JSON(200, gin.H{
		"message": "Video deleted",
	})
}

func addVideos(c *gin.Context, pm *controllers.PlayersManager) {
	TwitchUserContext, _ := c.Get("TwitchUser")
	userInfo, _ := TwitchUserContext.(twitch.UserInfo)
	userContext, _ := c.Get("User")
	user, _ := userContext.(models.Users)

	videoAddRequest := models.VideoAddRequest{}
	err := c.BindJSON(&videoAddRequest)

	if err != nil {
		c.JSON(400, gin.H{
			"error": "Failed to bind video add request",
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

	songRequest := models.SongRequest{
		YoutubeID: videoInfo.Id,
		Title:     videoInfo.Snippet.Title,
		Duration:  videoInfo.ContentDetails.Duration,
		Thumbnail: videoInfo.Snippet.Thumbnails.Default.Url,
		Type:      videoAddRequest.Type,
	}

	newVideo, err := controllers.InsertRequestInDatabase(songRequest, userInfo.ID)
	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	conn := pm.GetConnFromTwitchId(user.TwitchId)

	if conn != nil {
		for _, cn := range conn {
			cn.WriteJSON(newVideo)
		}
	}

	c.JSON(200, gin.H{
		"message": "Video added",
	})
}

func setCurrentPlaying(c *gin.Context) {
	userContext, _ := c.Get("User")
	user, _ := userContext.(models.Users)

	currentlyPlayingVideo := models.UsersVideos{}
	err := c.BindJSON(&currentlyPlayingVideo)

	if err != nil {
		c.JSON(400, gin.H{
			"error": "Failed to bind current playing request",
		})
		return
	}

	valkeyClient := utils.GetValkeyClient()
	jsonPayload, _ := json.Marshal(currentlyPlayingVideo)
	pub := valkeyClient.Publish(context.Background(), "player:"+user.TwitchId+":currently_playing", jsonPayload)
	if pub.Err() != nil {
		fmt.Println("Failed to publish currently playing video event", pub.Err())
		c.JSON(500, gin.H{
			"error": "Failed to publish currently playing video event",
		})
		return
	}
}
