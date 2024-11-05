package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Yaon-C2H8N2/bahclePlayer/models"
	"github.com/Yaon-C2H8N2/bahclePlayer/models/songRequests"
	"github.com/Yaon-C2H8N2/bahclePlayer/models/twitch"
	"github.com/Yaon-C2H8N2/bahclePlayer/utils"
	"github.com/gorilla/websocket"
	"regexp"
)

type NotificationHandler struct {
	handlers       map[string]func([]byte)
	apiWrapper     *ApiWrapper
	requestManager *songRequests.RequestManager
	token          string
	conn           *websocket.Conn
}

func GetNotificationHandler(apiWrapper *ApiWrapper, token string, conn *websocket.Conn) *NotificationHandler {
	handler := &NotificationHandler{
		handlers:       make(map[string]func([]byte)),
		apiWrapper:     apiWrapper,
		requestManager: songRequests.GetRequestManager(),
		token:          token,
		conn:           conn,
	}

	handler.handlers["channel.channel_points_custom_reward_redemption.add"] = handler.handleChannelPointsCustomRewardRedemptionAdd
	handler.handlers["channel.poll.end"] = handler.handleChannelPollEnd

	return handler
}

func (nh *NotificationHandler) Handle(notificationBytes []byte) {
	var notification = &twitch.NotificationMessage{}
	err := json.Unmarshal(notificationBytes, notification)
	if err != nil {
		fmt.Println("Failed to unmarshal notification")
		return
	}

	eventBytes, _ := json.Marshal(notification.Payload.Event)
	if _, ok := nh.handlers[notification.Payload.Subscription.Type]; ok {
		nh.handlers[notification.Payload.Subscription.Type](eventBytes)
	}
}

func (nh *NotificationHandler) handleChannelPointsCustomRewardRedemptionAdd(eventBytes []byte) {
	var redemptionEvent = &twitch.ChannelPointsCustomRewardRedemptionAddEvent{}
	err := json.Unmarshal(eventBytes, redemptionEvent)
	if err != nil {
		fmt.Println("Failed to unmarshal redemption event")
		return
	}

	conn := utils.GetConnection()
	defer conn.Close(context.Background())

	sql := `
			SELECT config, value
			FROM users_config
			WHERE user_id = $1
		`
	rows := utils.DoRequest(conn, sql, redemptionEvent.BroadcasterUserId)
	var config []struct{ Config, Value string }
	for rows.Next() {
		var result struct{ Config, Value string }
		err = rows.Scan(&result.Config, &result.Value)
		if err != nil {
			fmt.Println("Failed to get user config")
			return
		}
		config = append(config, result)
	}
	if len(config) == 0 {
		fmt.Println("Failed to get user config")
		return
	}

	found := false
	for _, c := range config {
		if (c.Config == "PLAYLIST_REDEMPTION" || c.Config == "QUEUE_REDEMPTION") && c.Value == redemptionEvent.Reward.Id {
			found = true
			break
		}
	}
	if !found {
		fmt.Println("Redemption reward not found in user config")
		return
	}

	message := redemptionEvent.UserInput
	re := regexp.MustCompile(`(?:youtu\.be/|youtube\.com/(?:embed/|v/|watch\?v=|watch\?.+&v=))([^&\n?#]+)`)
	match := re.FindStringSubmatch(message)
	youtubeId := ""
	if len(match) > 1 {
		youtubeId = match[1]
	}

	if youtubeId == "" {
		fmt.Println("Failed to extract youtube id from message")
		err = nh.apiWrapper.UpdateRedemptionStatus(nh.token, redemptionEvent.Id, redemptionEvent.BroadcasterUserId, redemptionEvent.Reward.Id, "CANCELED")
		if err != nil {
			fmt.Println("Failed to update redemption status")
		}
		return
	}
	youtubeResults, err := GetVideoInfo(youtubeId)
	if err != nil {
		fmt.Println("Failed to get video info")
		err = nh.apiWrapper.UpdateRedemptionStatus(nh.token, redemptionEvent.Id, redemptionEvent.BroadcasterUserId, redemptionEvent.Reward.Id, "CANCELED")
		return
	}
	video := youtubeResults.Items[0]

	var songRequest = songRequests.SongRequest{}
	songRequest.TwitchRedemptionID = redemptionEvent.Id
	songRequest.TwitchRewardID = redemptionEvent.Reward.Id
	songRequest.YoutubeID = youtubeId
	songRequest.Title = video.Snippet.Title
	songRequest.Channel = video.Snippet.ChannelTitle
	songRequest.Duration = video.ContentDetails.Duration
	songRequest.Thumbnail = video.Snippet.Thumbnails.Default.Url

	//TODO : create settings for when to add the track to the playlist (immediately, after poll, ...)
	pollTitle := fmt.Sprintf("Add the current track to playlist ?")
	twitchPollId, err := nh.apiWrapper.CreatePoll(nh.token, redemptionEvent.BroadcasterUserId, pollTitle, []string{"Yes", "No"}, 60)
	if err != nil || twitchPollId == "" {
		if err != nil {
			fmt.Printf("Failed to create poll : %s\n", err.Error())
		} else {
			fmt.Println("Failed to create poll : empty poll id")
		}
		err = nh.apiWrapper.UpdateRedemptionStatus(nh.token, redemptionEvent.Id, redemptionEvent.BroadcasterUserId, redemptionEvent.Reward.Id, "CANCELED")
		return
	}
	songRequest.TwitchPollID = twitchPollId
	nh.requestManager.AddRequest(songRequest)
}

func (nh *NotificationHandler) handleChannelPollEnd(eventBytes []byte) {
	var pollEndEvent = &twitch.ChannelPollEndEvent{}
	err := json.Unmarshal(eventBytes, pollEndEvent)
	if err != nil {
		fmt.Println("Failed to unmarshal poll end event")
		return
	}

	if pollEndEvent.Status != "completed" {
		return
	}
	songRequest := nh.requestManager.GetRequest(pollEndEvent.Id)
	if songRequest.TwitchPollID == "" {
		fmt.Println("Failed to get song request from poll id")
		return
	}

	maxVotes := 0
	maxChoice := ""
	newStatus := "CANCELED"
	for _, choice := range pollEndEvent.Choices {
		if choice.Votes > maxVotes {
			maxVotes = choice.Votes
			maxChoice = choice.Title
		}
	}

	if maxChoice == "Yes" && maxVotes > 0 {
		newStatus = "FULFILLED"

		conn := utils.GetConnection()
		defer conn.Close(context.Background())

		sql := `
			SELECT config, value
			FROM users_config
			WHERE user_id = $1
		`
		rows := utils.DoRequest(conn, sql, pollEndEvent.BroadcasterUserId)
		var config []struct{ Config, Value string }
		for rows.Next() {
			var result struct{ Config, Value string }
			err = rows.Scan(&result.Config, &result.Value)
			if err != nil {
				fmt.Println("Failed to get user config")
				return
			}
			config = append(config, result)
		}
		if len(config) == 0 {
			fmt.Println("Failed to get user config")
			return
		}

		requestType := ""
		for _, c := range config {
			if (c.Config == "PLAYLIST_REDEMPTION" || c.Config == "QUEUE_REDEMPTION") && c.Value == songRequest.TwitchRewardID {
				if c.Config == "PLAYLIST_REDEMPTION" {
					requestType = "PLAYLIST"
				} else {
					requestType = "QUEUE"
				}
				break
			}
		}
		if requestType == "" {
			fmt.Println("Redemption reward not found in user config")
			return
		}

		var newVideo = models.UsersVideos{}
		sql = `
				INSERT INTO users_videos(user_id, youtube_id, url, title, duration, type, thumbnail_url, added_by)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
				RETURNING *;
			`
		rows = utils.DoRequest(
			conn,
			sql,
			pollEndEvent.BroadcasterUserId,
			songRequest.YoutubeID,
			"https://www.youtube.com/watch?v="+songRequest.YoutubeID,
			songRequest.Title,
			songRequest.Duration,
			requestType,
			songRequest.Thumbnail,
			"twitch", //TODO : get added_by from user
		)
		if !rows.Next() {
			fmt.Println("Failed to insert video into database")
			return
		}
		rows.Scan(&newVideo.VideoId, &newVideo.UserId, &newVideo.YoutubeId, &newVideo.Url, &newVideo.Title, &newVideo.Duration, &newVideo.Type, &newVideo.CreatedAt, &newVideo.ThumbnailUrl, &newVideo.AddedBy)

		err = nh.conn.WriteJSON(newVideo)
		if err != nil {
			fmt.Println("Failed to send song request to player")
			return
		}
	}
	err = nh.apiWrapper.UpdateRedemptionStatus(nh.token, songRequest.TwitchRedemptionID, pollEndEvent.BroadcasterUserId, songRequest.TwitchRewardID, newStatus)
	if err != nil {
		fmt.Println("Failed to update redemption status")
		return
	}
}
