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
	//TODO : Get PLAYLIST_REDEMPTION and QUEUE_REDEMPTION values from database

	//TODO : if redemptionEvent.Reward.Id not equal to PLAYLIST_REDEMPTION or QUEUE_REDEMPTION, return

	message := redemptionEvent.UserInput
	youtubeIdRegexp := regexp.MustCompile(`(https?://)?(www\.)?(youtube|youtu|youtube-nocookie)\.(com|be)/((watch\?v=|embed/|v/|e/|u/\w+/|v=|\?v=)?)([^#&?]{11})`)
	youtubeId := youtubeIdRegexp.FindAllString(message, -1)[6]
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

	pollTitle := fmt.Sprintf("Should we play %s by %s?", songRequest.Title, songRequest.Channel)
	twitchPollId, err := nh.apiWrapper.CreatePoll(nh.token, redemptionEvent.BroadcasterUserId, pollTitle, []string{"Yes", "No"}, 60)
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

	songRequest := nh.requestManager.GetRequest(pollEndEvent.Id)
	if songRequest.TwitchPollID == "" {
		fmt.Println("Failed to get song request from poll id")
		return
	}

	maxVotes := 0
	maxChoice := ""
	newStatus := "CANCELED"
	for _, choice := range pollEndEvent.Choices {
		if choice.BitsVotes+choice.ChannelPointsVotes > maxVotes {
			maxVotes = choice.BitsVotes + choice.ChannelPointsVotes
			maxChoice = choice.Title
		}
	}
	if maxChoice == "Yes" {
		newStatus = "FULFILLED"

		conn := utils.GetConnection()
		defer conn.Close(context.Background())

		//TODO : compare redemptionRequest id with PLAYLIST_REDEMPTION and QUEUE_REDEMPTION values from database to get proper type
		requestType := "PLAYLIST"

		var newVideo = &models.UsersVideos{}
		sql := `
				INSERT INTO users_videos(user_id, youtube_id, url, title, duration, type, thumbnail_url, added_by)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
				RETURNING *;
			`
		rows := utils.DoRequest(
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
		rows.Scan(newVideo)

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
