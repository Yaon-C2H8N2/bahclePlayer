package songRequests

import (
	"fmt"
	"github.com/Yaon-C2H8N2/bahclePlayer/internal/models"
	"github.com/Yaon-C2H8N2/bahclePlayer/pkg/utils"
)

type RequestManager struct {
	requests map[string]SongRequest
}

func GetRequestManager() *RequestManager {
	return &RequestManager{
		requests: make(map[string]SongRequest),
	}
}

func (rm *RequestManager) AddRequest(request SongRequest) {
	rm.requests[request.TwitchPollID] = request
}

func (rm *RequestManager) GetRequest(pollId string) SongRequest {
	var request, ok = rm.requests[pollId]
	if ok {
		return request
	}
	return SongRequest{}
}

func (rm *RequestManager) RemoveRequest(pollId string) {
	delete(rm.requests, pollId)
}

func (rm *RequestManager) GetAllRequests() map[string]SongRequest {
	return rm.requests
}

func (rm *RequestManager) ClearRequests() {
	rm.requests = make(map[string]SongRequest)
}

func InsertRequestInDatabase(songRequest SongRequest, broadcasterUserId string) (models.UsersVideos, error) {
	conn := utils.GetConnection()
	defer conn.Release()

	var newVideo = models.UsersVideos{}
	sql := `
				INSERT INTO users_videos(user_id, youtube_id, url, title, duration, type, thumbnail_url, added_by)
				VALUES ((SELECT user_id FROM users WHERE twitch_id = $1), $2, $3, $4, $5, $6, $7, $8)
				RETURNING *;
			`
	rows := utils.DoRequest(
		conn,
		sql,
		broadcasterUserId,
		songRequest.YoutubeID,
		"https://www.youtube.com/watch?v="+songRequest.YoutubeID,
		songRequest.Title,
		songRequest.Duration,
		songRequest.Type,
		songRequest.Thumbnail,
		"twitch", //TODO : get added_by from user
	)
	if !rows.Next() {
		return newVideo, fmt.Errorf("Failed to insert video into database")
	}
	rows.Scan(&newVideo.VideoId, &newVideo.UserId, &newVideo.YoutubeId, &newVideo.Url, &newVideo.Title, &newVideo.Duration, &newVideo.Type, &newVideo.CreatedAt, &newVideo.ThumbnailUrl, &newVideo.AddedBy)
	return newVideo, nil
}
