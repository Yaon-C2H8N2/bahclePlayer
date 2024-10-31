package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/Yaon-C2H8N2/bahclePlayer/models/songRequests"
	"github.com/Yaon-C2H8N2/bahclePlayer/models/twitch"
	"regexp"
)

type NotificationHandler struct {
	handlers       map[string]func([]byte)
	apiWrapper     *ApiWrapper
	requestManager *songRequests.RequestManager
	token          string
}

func GetNotificationHandler(apiWrapper *ApiWrapper, token string) *NotificationHandler {
	handler := &NotificationHandler{
		handlers:       make(map[string]func([]byte)),
		apiWrapper:     apiWrapper,
		requestManager: songRequests.GetRequestManager(),
		token:          token,
	}

	handler.handlers["channel.channel_points_custom_reward_redemption.add"] = handler.handleChannelPointsCustomRewardRedemptionAdd

	return handler
}

func (nh *NotificationHandler) Handle(notification *twitch.NotificationMessage) {
	notificationBytes, _ := json.Marshal(notification)
	nh.handlers[notification.Payload.Subscription.Type](notificationBytes)
}

func (nh *NotificationHandler) handleChannelPointsCustomRewardRedemptionAdd(eventBytes []byte) {
	var redemptionEvent = &twitch.ChannelPointsCustomRewardRedemptionAddEvent{}
	err := json.Unmarshal(eventBytes, redemptionEvent)
	if err != nil {
		fmt.Println("Failed to unmarshal redemption event")
		return
	}

	message := redemptionEvent.UserInput
	youtubeIdRegexp := regexp.MustCompile(`(https?://)?(www\.)?(youtube|youtu|youtube-nocookie)\.(com|be)/((watch\?v=|embed/|v/|e/|u/\w+/|v=|\?v=)?)([^#&?]{11})`)
	youtubeId := youtubeIdRegexp.FindString(message)
	if youtubeId == "" {
		fmt.Println("Failed to extract youtube id from message")
		err = nh.apiWrapper.UpdateRedemptionStatus(nh.token, redemptionEvent.Id, redemptionEvent.BroadcasterUserId, redemptionEvent.Reward.Id, "CANCELED")
		return
	}

	var songRequest = songRequests.SongRequest{}
	songRequest.TwitchRedemptionID = redemptionEvent.Id
	songRequest.TwitchRewardID = redemptionEvent.Reward.Id
	songRequest.YoutubeID = youtubeId
	songRequest.Title = ""     //TODO: get video title
	songRequest.Channel = ""   //TODO: get video channel
	songRequest.Duration = 0   //TODO: get video duration
	songRequest.Thumbnail = "" //TODO: get video base64 thumbnail or url

	pollTitle := fmt.Sprintf("Should we play %s by %s?", songRequest.Title, songRequest.Channel)
	twitchPollId, err := nh.apiWrapper.CreatePoll(nh.token, redemptionEvent.BroadcasterUserId, pollTitle, []string{"Yes", "No"}, 60)
	songRequest.TwitchPollID = twitchPollId
	nh.requestManager.AddRequest(songRequest)
}

func (nh *NotificationHandler) handleChannelPollEnd(token string, eventBytes []byte) {
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
		//TODO: send song request to player if valid
		newStatus = "FULFILLED"
	}
	err = nh.apiWrapper.UpdateRedemptionStatus(token, songRequest.TwitchRedemptionID, pollEndEvent.BroadcasterUserId, songRequest.TwitchRewardID, newStatus)
	if err != nil {
		fmt.Println("Failed to update redemption status")
		return
	}
}
