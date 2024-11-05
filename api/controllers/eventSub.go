package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/Yaon-C2H8N2/bahclePlayer/models/twitch"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type EventSub struct {
	onEvent    map[string]func(event twitch.NotificationMessage)
	onStarted  func()
	sessionId  string
	apiWrapper *ApiWrapper
}

func GetEventSub(apiWrapper *ApiWrapper) *EventSub {
	var newEventSub = &EventSub{
		onEvent: make(map[string]func(event twitch.NotificationMessage)),
	}

	newEventSub.apiWrapper = apiWrapper
	return newEventSub
}

func (es *EventSub) Start() {
	es.listenToMessages()
}

func (es *EventSub) OnEvent(token string, callback func(event twitch.NotificationMessage)) {
	userId, _ := es.apiWrapper.GetUserInfoFromToken(token)

	es.onEvent[userId.ID] = callback
}

func (es *EventSub) OnStarted(callback func()) {
	es.onStarted = callback
}

func (es *EventSub) IsStarted() bool {
	return es.sessionId != ""
}

func (es *EventSub) DropAllSubscriptions(userToken string) {
	//TODO: Implement
}

func (es *EventSub) SubscribeToMessageEvents(userToken string) {
	twitchUrl := "https://api.twitch.tv/helix/eventsub/subscriptions"
	broadcasterId, err := es.apiWrapper.GetUserInfoFromToken(userToken)
	if err != nil {
		panic(err)
	}

	var data = twitch.SubscriptionRequest{
		Type:    "channel.chat.message",
		Version: "1",
		Condition: twitch.Condition{
			BroadcasterUserId: broadcasterId.ID,
			UserId:            broadcasterId.ID,
		},
		Transport: twitch.Transport{
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

	subcriptionResponse := &twitch.SubscriptionResponse{}
	err = json.Unmarshal(body, subcriptionResponse)
	if err != nil {
		fmt.Println("Error unmarshalling response:", err)
		return
	}
	fmt.Println("Total subscriptions count:", subcriptionResponse.Total)
}

func (es *EventSub) SubscribeToRedemptionEvents(userToken string) {
	twitchUrl := "https://api.twitch.tv/helix/eventsub/subscriptions"
	broadcasterId, err := es.apiWrapper.GetUserInfoFromToken(userToken)
	if err != nil {
		panic(err)
	}

	var data = twitch.SubscriptionRequest{
		Type:    "channel.channel_points_custom_reward_redemption.add",
		Version: "1",
		Condition: twitch.Condition{
			BroadcasterUserId: broadcasterId.ID,
		},
		Transport: twitch.Transport{
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

	subcriptionResponse := &twitch.SubscriptionResponse{}
	err = json.Unmarshal(body, subcriptionResponse)
	if err != nil {
		fmt.Println("Error unmarshalling response:", err)
		return
	}
	fmt.Println("Total subscriptions count:", subcriptionResponse.Total)
}

func (es *EventSub) SubscribeToPollEvents(userToken string) {
	twitchUrl := "https://api.twitch.tv/helix/eventsub/subscriptions"
	broadcasterId, err := es.apiWrapper.GetUserInfoFromToken(userToken)
	if err != nil {
		panic(err)
	}

	var data = twitch.SubscriptionRequest{
		Type:    "channel.poll.end",
		Version: "1",
		Condition: twitch.Condition{
			BroadcasterUserId: broadcasterId.ID,
		},
		Transport: twitch.Transport{
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

	subcriptionResponse := &twitch.SubscriptionResponse{}
	err = json.Unmarshal(body, subcriptionResponse)
	if err != nil {
		fmt.Println("Error unmarshalling response:", err)
		return
	}
	fmt.Println("Total subscriptions count:", subcriptionResponse.Total)
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

			var message = &twitch.BaseMessage{}
			err = json.Unmarshal(messageBytes, message)
			if err != nil {
				log.Printf("err: %s", messageBytes)
				panic(err)
			}

			switch message.Metadata.MessageType {
			case "session_welcome":
				var welcomeMessage = &twitch.WelcomeMessage{}
				err = json.Unmarshal(messageBytes, welcomeMessage)
				if err != nil {
					log.Printf("err: %s", messageBytes)
					panic(err)
				}

				es.sessionId = welcomeMessage.Payload.Session.Id
				go es.onStarted()
				break
			case "notification":
				var notificationMessage = &twitch.NotificationMessage{}
				err = json.Unmarshal(messageBytes, notificationMessage)
				if err != nil {
					log.Printf("err: %s", messageBytes)
					panic(err)
				}
				go es.onEvent[notificationMessage.Payload.Subscription.Condition.BroadcasterUserId](*notificationMessage)
				break
			}
		}
	}()
	fmt.Println("Listening to messages")
}
