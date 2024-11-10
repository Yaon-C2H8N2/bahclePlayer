package endpoints

import (
	"context"
	"github.com/Yaon-C2H8N2/bahclePlayer/controllers"
	"github.com/Yaon-C2H8N2/bahclePlayer/models"
	"github.com/Yaon-C2H8N2/bahclePlayer/utils"
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
}

func getPlaylistAndQueue(c *gin.Context, aw *controllers.ApiWrapper) {
	token := c.Request.Header.Get("Authorization")
	token = token[7:]
	if token == "" {
		c.JSON(400, gin.H{
			"error": "missing access_token",
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

	conn := utils.GetConnection()
	defer conn.Close(context.Background())

	sql := `
			SELECT users_videos.*
			FROM users_videos
			JOIN users ON users_videos.user_id = users.user_id
			WHERE users.twitch_id = $1
			ORDER BY users_videos.created_at DESC
		`
	rows := utils.DoRequest(conn, sql, userInfo.ID)
	var results []models.UsersVideos
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

func deleteVideo(c *gin.Context, aw *controllers.ApiWrapper) {
	token := c.Request.Header.Get("Authorization")
	token = token[7:]
	if token == "" {
		c.JSON(400, gin.H{
			"error": "missing access_token",
		})
		return
	}

	videoID := c.Query("video_id")
	if videoID == "" {
		c.JSON(400, gin.H{
			"error": "missing video_id",
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

	conn := utils.GetConnection()
	defer conn.Close(context.Background())

	sql := `
			DELETE FROM users_videos
			WHERE user_id = (SELECT user_id FROM users WHERE twitch_id = $1) AND video_id = $2
		`
	utils.DoRequest(conn, sql, userInfo.ID, videoID)

	c.JSON(200, gin.H{
		"message": "Video deleted",
	})
}
