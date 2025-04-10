package services

import (
	"github.com/Yaon-C2H8N2/bahclePlayer/internal/controllers"
	"github.com/Yaon-C2H8N2/bahclePlayer/internal/models"
	"github.com/Yaon-C2H8N2/bahclePlayer/internal/models/twitch"
	"github.com/Yaon-C2H8N2/bahclePlayer/pkg/utils"
	"github.com/gin-gonic/gin"
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
		results = append(results, result)
	}
	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
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
