package twitch

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type EventSub struct {
	onEvent   map[string]func(event any)
	onStarted func()
	sessionId string
	appToken  string
}

func GetEventSub() *EventSub {
	var newEventSub = &EventSub{
		onEvent: make(map[string]func(event any)),
	}
	newEventSub.setAppToken()
	return newEventSub
}

func (es *EventSub) Start() {
	es.listenToMessages()
}

func (es *EventSub) OnEvent(token string, callback func(event any)) {
	userId, _ := es.getBroadcasterIdFromToken(token)

	es.onEvent[userId] = callback
}

func (es *EventSub) OnStarted(callback func()) {
	es.onStarted = callback
}

func (es *EventSub) IsStarted() bool {
	return es.sessionId != ""
}

func (es *EventSub) DropAllSubscriptions(userToken string) {

}

func (es *EventSub) SubscribeToMessageEvents(userToken string) {
	twitchUrl := "https://api.twitch.tv/helix/eventsub/subscriptions"
	broadcasterId, err := es.getBroadcasterIdFromToken(userToken)
	if err != nil {
		panic(err)
	}

	var data = subscriptionRequest{
		Type:    "channel.chat.message",
		Version: "1",
		Condition: condition{
			BroadcasterUserId: broadcasterId,
			UserId:            broadcasterId,
		},
		Transport: transport{
			Method:    "websocket",
			SessionId: es.sessionId,
		},
	}

	httpClient := &http.Client{}
	bytes, _ := json.Marshal(data)
	fmt.Println("Request body:", string(bytes))
	req, err := http.NewRequest("POST", twitchUrl, strings.NewReader(string(bytes)))
	if err != nil {
		panic(err)
	}

	req.Header.Add("Authorization", "Bearer "+userToken)
	req.Header.Add("Client-Id", os.Getenv("TWITCH_CLIENT_ID"))
	req.Header.Add("Content-Type", "application/json")

	res, err := httpClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return
	}

	subcriptionResponse := &subcriptionResponse{}
	err = json.Unmarshal(body, subcriptionResponse)
	if err != nil {
		fmt.Println("Error unmarshalling response:", err)
		return
	}
	fmt.Println("Total subscriptions count:", subcriptionResponse.Total)
}

func (es *EventSub) setAppToken() {
	twitchUrl := "https://id.twitch.tv/oauth2/token"

	request := &appTokenRequest{
		ClientId:     os.Getenv("TWITCH_CLIENT_ID"),
		ClientSecret: os.Getenv("TWITCH_CLIENT_SECRET"),
		GrantType:    "client_credentials",
	}
	twitchUrl += "?client_id=" + request.ClientId + "&client_secret=" + request.ClientSecret + "&grant_type=" + request.GrantType

	httpClient := &http.Client{}
	req, err := http.NewRequest("POST", twitchUrl, nil)
	if err != nil {
		panic(err)
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := httpClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return
	}

	var tokenResponse = &tokenResponse{}
	err = json.Unmarshal(body, tokenResponse)
	if err != nil {
		fmt.Println("Error unmarshalling response:", err)
		return
	}

	es.appToken = tokenResponse.AccessToken
}

func (es *EventSub) getBroadcasterIdFromToken(userToken string) (string, error) {
	twitchUrl := "https://api.twitch.tv/helix/users"

	httpClient := &http.Client{}
	req, err := http.NewRequest("GET", twitchUrl, nil)
	if err != nil {
		return "", err
	}

	req.Header.Add("Authorization", "Bearer "+userToken)
	req.Header.Add("Client-Id", os.Getenv("TWITCH_CLIENT_ID"))

	res, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	var userInfoResponse = &userInfoResponse{}
	err = json.Unmarshal(body, userInfoResponse)
	if err != nil {
		return "", err
	}

	return userInfoResponse.Data[0].ID, nil
}

func (es *EventSub) listenToMessages() {
	//https://github.com/gorilla/websocket/blob/main/examples/echo/client.go
	conn, _, err := websocket.DefaultDialer.Dial("wss://eventsub.wss.twitch.tv/ws", nil)

	if err != nil {
		panic(err)
	}

	go func() {
		defer conn.Close()
		for {
			_, messageBytes, err := conn.ReadMessage()
			if err != nil {
				log.Printf("err: %s", messageBytes)
				panic(err)
			}

			var message = &BaseMessage{}
			err = json.Unmarshal(messageBytes, message)
			if err != nil {
				log.Printf("err: %s", messageBytes)
				panic(err)
			}

			switch message.Metadata.MessageType {
			case "session_welcome":
				var welcomeMessage = &WelcomeMessage{}
				err = json.Unmarshal(messageBytes, welcomeMessage)
				if err != nil {
					log.Printf("err: %s", messageBytes)
					panic(err)
				}

				es.sessionId = welcomeMessage.Payload.Session.Id
				go es.onStarted()
				break
			case "notification":
				var notificationMessage = &NotificationMessage{}
				err = json.Unmarshal(messageBytes, notificationMessage)
				if err != nil {
					log.Printf("err: %s", messageBytes)
					panic(err)
				}
				go es.onEvent[notificationMessage.Payload.Subscription.Condition.BroadcasterUserId](notificationMessage)
				break
			}
		}
	}()
	fmt.Println("Listening to messages")
}
