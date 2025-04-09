package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/Yaon-C2H8N2/bahclePlayer/internal/models"
	"github.com/Yaon-C2H8N2/bahclePlayer/internal/models/songRequests"
	"github.com/Yaon-C2H8N2/bahclePlayer/internal/models/twitch"
	"github.com/Yaon-C2H8N2/bahclePlayer/pkg/utils"
	"regexp"
)

type NotificationHandler struct {
	handlers       map[string]func([]byte)
	apiWrapper     *ApiWrapper
	requestManager *songRequests.RequestManager
	token          string
	onNewVideo     func(video models.UsersVideos)
}

func (nh *NotificationHandler) OnNewVideo(callback func(event models.UsersVideos)) func() {
	nh.onNewVideo = callback

	return func() {
		nh.onNewVideo = nil
	}
}

func GetNotificationHandler(apiWrapper *ApiWrapper, token string) *NotificationHandler {
	handler := &NotificationHandler{
		handlers:       make(map[string]func([]byte)),
		apiWrapper:     apiWrapper,
		requestManager: songRequests.GetRequestManager(),
		token:          token,
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
	defer conn.Release()

	sql := `
			SELECT config, value
			FROM users_config
			JOIN users ON users_config.user_id = users.user_id
			WHERE users.twitch_id = $1
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

	var songRequest = songRequests.SongRequest{}
	for _, c := range config {
		if c.Config == "PLAYLIST_REDEMPTION" && c.Value == redemptionEvent.Reward.Id {
			songRequest.Type = "PLAYLIST"
			break
		}
		if c.Config == "QUEUE_REDEMPTION" && c.Value == redemptionEvent.Reward.Id {
			songRequest.Type = "QUEUE"
			break
		}
	}

	for _, c := range config {
		if c.Config == (songRequest.Type + "_METHOD") {
			songRequest.Method = c.Value
			break
		}
	}
	if songRequest.Method == "" {
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
	fmt.Printf("New video request : {twitch_id: %s, youtube_id: %s, method: %s}\n", redemptionEvent.BroadcasterUserId, youtubeId, songRequest.Method)
	youtubeResults, err := GetVideoInfo(youtubeId)
	if err != nil {
		fmt.Println("Failed to get video info")
		err = nh.apiWrapper.UpdateRedemptionStatus(nh.token, redemptionEvent.Id, redemptionEvent.BroadcasterUserId, redemptionEvent.Reward.Id, "CANCELED")
		return
	}
	video := youtubeResults.Items[0]

	songRequest.TwitchRedemptionID = redemptionEvent.Id
	songRequest.TwitchRewardID = redemptionEvent.Reward.Id
	songRequest.YoutubeID = youtubeId
	songRequest.Title = video.Snippet.Title
	songRequest.Channel = video.Snippet.ChannelTitle
	songRequest.Duration = video.ContentDetails.Duration
	songRequest.Thumbnail = video.Snippet.Thumbnails.Default.Url

	if songRequest.Method == "POLL" {
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
	} else if songRequest.Method == "DIRECT" {
		newVideo, err := songRequests.InsertRequestInDatabase(songRequest, redemptionEvent.BroadcasterUserId)
		if err != nil {
			fmt.Println("Failed to insert video into database")
			return
		}

		if nh.onNewVideo != nil {
			nh.onNewVideo(newVideo)
		}
	}
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

		newVideo, err := songRequests.InsertRequestInDatabase(songRequest, pollEndEvent.BroadcasterUserId)
		if err != nil {
			fmt.Println("Failed to insert video into database")
			return
		}

		if nh.onNewVideo != nil {
			nh.onNewVideo(newVideo)
		}
	}
	err = nh.apiWrapper.UpdateRedemptionStatus(nh.token, songRequest.TwitchRedemptionID, pollEndEvent.BroadcasterUserId, songRequest.TwitchRewardID, newStatus)
	if err != nil {
		fmt.Println("Failed to update redemption status")
		return
	}
}
