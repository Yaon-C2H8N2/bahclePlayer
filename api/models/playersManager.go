package models

import (
	"encoding/json"
	"fmt"
	"github.com/Yaon-C2H8N2/bahclePlayer/models/songRequests"
	"github.com/Yaon-C2H8N2/bahclePlayer/models/twitch"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
	"os"
	"regexp"
	"sync"
)

type PlayersManager struct {
	mutex      *sync.Mutex
	clients    map[string]*websocket.Conn
	upgrader   websocket.Upgrader
	eventSub   *twitch.EventSub
	apiWrapper *twitch.ApiWrapper
}

func DefaultPlayersManager() *PlayersManager {
	apiWrapper := twitch.GetApiWrapper()
	appToken, err := twitch.RequestAppToken(os.Getenv("TWITCH_CLIENT_ID"), os.Getenv("TWITCH_CLIENT_SECRET"))
	if err != nil {
		panic(err)
	}
	apiWrapper.SetAppToken(appToken)
	apiWrapper.SetClientId(os.Getenv("TWITCH_CLIENT_ID"))

	return &PlayersManager{
		mutex:   &sync.Mutex{},
		clients: make(map[string]*websocket.Conn),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		eventSub:   twitch.GetEventSub(apiWrapper),
		apiWrapper: apiWrapper,
	}
}

func (pm *PlayersManager) CreatePlayer(c *gin.Context) {
	pm.mutex.Lock()
	if _, ok := pm.clients[c.Query("access_token")]; !ok {
		conn, err := pm.upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			c.JSON(500, gin.H{
				"message": "Failed to upgrade connection",
				"error":   err.Error(),
			})
			return
		}

		token := c.Query("access_token")
		pm.clients[token] = conn

		go pm.mainLoop(token)
	} else {
		c.JSON(400, gin.H{
			"message": "Player already exists",
		})
	}
	pm.mutex.Unlock()
}

func (pm *PlayersManager) mainLoop(token string) {
	conn := pm.clients[token]

	requestManager := songRequests.GetRequestManager()

	pm.eventSub.OnEvent(token, func(event twitch.NotificationMessage) {
		eventBytes, _ := json.Marshal(event.Payload.Event)

		switch event.Payload.Subscription.Type {
		case "channel.channel_points_custom_reward_redemption.add":
			var redemptionEvent = &twitch.ChannelPointsCustomRewardRedemptionAddEvent{}
			err := json.Unmarshal(eventBytes, redemptionEvent)
			if err != nil {
				fmt.Println("Failed to unmarshal redemption event")
				break
			}

			message := redemptionEvent.UserInput
			youtubeIdRegexp := regexp.MustCompile(`(https?://)?(www\.)?(youtube|youtu|youtube-nocookie)\.(com|be)/((watch\?v=|embed/|v/|e/|u/\w+/|v=|\?v=)?)([^#&?]{11})`)
			youtubeId := youtubeIdRegexp.FindString(message)
			if youtubeId == "" {
				fmt.Println("Failed to extract youtube id from message")
				//TODO: request twitch to cancel redemption
				break
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
			twitchPollId, err := pm.apiWrapper.CreatePoll(token, redemptionEvent.BroadcasterUserId, pollTitle, []string{"Yes", "No"}, 60)
			songRequest.TwitchPollID = twitchPollId //TODO: get poll id
			requestManager.AddRequest(songRequest)
			break
		case "channel.poll.end":
			//TODO: get poll id
			//TODO: if poll id is in request manager, get and remove song request
			//TODO: check if poll result is valid
			//TODO: send song request to player if valid
			//TODO: set redemption as completed if valid
			//TODO: set redemption as canceled if invalid
		}
	})

	if pm.eventSub.IsStarted() {
		pm.eventSub.SubscribeToMessageEvents(token)
	} else {
		pm.eventSub.OnStarted(func() {
			pm.eventSub.SubscribeToMessageEvents(token)
		})
		pm.eventSub.Start()
	}

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Client disconnected")
			conn.Close()

			pm.mutex.Lock()
			delete(pm.clients, token)
			pm.mutex.Unlock()

			pm.eventSub.DropAllSubscriptions(token)

			break
		}
	}
}
