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
	onEvent       func(event any)
	onStarted     func()
	sessionId     string
	token         string
	broadcasterId string
}

func GetEventSub(token string) *EventSub {
	var newEventSub = &EventSub{
		token: token,
	}
	newEventSub.setBroadcasterIdFromToken()
	return newEventSub
}

func (es *EventSub) Start() {
	es.listenToMessages()
}

func (es *EventSub) OnEvent(callback func(event any)) {
	es.onEvent = callback
}

func (es *EventSub) OnStarted(callback func()) {
	es.onStarted = callback
}

func (es *EventSub) SubscribeToMessageEvents() {
	twitchUrl := "https://api.twitch.tv/helix/eventsub/subscriptions"

	var data = subscriptionRequest{
		Type:    "channel.chat.message",
		Version: "1",
		Condition: condition{
			BroadcasterUserId: es.broadcasterId,
			UserId:            es.broadcasterId,
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

	req.Header.Add("Authorization", "Bearer "+es.token)
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

	fmt.Println("Response body:", string(body))
}

func (es *EventSub) setBroadcasterIdFromToken() {
	twitchUrl := "https://api.twitch.tv/helix/users"

	httpClient := &http.Client{}
	req, err := http.NewRequest("GET", twitchUrl, nil)
	if err != nil {
		panic(err)
	}

	req.Header.Add("Authorization", "Bearer "+es.token)
	req.Header.Add("Client-Id", os.Getenv("TWITCH_CLIENT_ID"))

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
	var userInfoResponse = &userInfoResponse{}
	err = json.Unmarshal(body, userInfoResponse)
	if err != nil {
		fmt.Println("Error unmarshalling response:", err)
		return
	}
	es.broadcasterId = userInfoResponse.Data[0].ID
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
				es.onStarted()
				break
			case "notification":
				var notificationMessage = &NotificationMessage{}
				err = json.Unmarshal(messageBytes, notificationMessage)
				if err != nil {
					log.Printf("err: %s", messageBytes)
					panic(err)
				}
				es.onEvent(notificationMessage)
				break
			}
		}
	}()
}
