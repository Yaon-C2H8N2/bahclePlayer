package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Yaon-C2H8N2/bahclePlayer/models"
	"github.com/Yaon-C2H8N2/bahclePlayer/models/twitch"
	"github.com/Yaon-C2H8N2/bahclePlayer/utils"
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

func (es *EventSub) OnEvent(token string, callback func(event twitch.NotificationMessage)) func() {
	userId, _ := es.apiWrapper.GetUserInfoFromToken(token)

	es.onEvent[userId.ID] = callback

	return func() {
		delete(es.onEvent, userId.ID)
	}
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

func (es *EventSub) InitSubscriptions(userToken string) {
	var err error
	err = es.subscribeToMessageEvents(userToken)
	if err != nil {
		fmt.Printf("Error subscribing to message events with token %s: %s\n", userToken, err)
	}
	err = es.subscribeToRedemptionEvents(userToken)
	if err != nil {
		fmt.Printf("Error subscribing to redemption events with token %s: %s\n", userToken, err)
	}
	err = es.subscribeToPollEvents(userToken)
	if err != nil {
		fmt.Printf("Error subscribing to poll events with token %s: %s\n", userToken, err)
	}
}

func (es *EventSub) InitForAllUsers() {
	conn := utils.GetConnection()
	defer conn.Close(context.Background())

	sql := `
			SELECT * FROM users
		`
	rows := utils.DoRequest(conn, sql)
	var users []models.Users
	for rows.Next() {
		var user models.Users
		rows.Scan(&user.UserId, &user.Username, &user.TwitchId, &user.Token, &user.TokenCreatedAt)
		users = append(users, user)
	}

	for _, user := range users {
		if user.Token == "" {
			continue
		}
		fmt.Printf("Initializing subscriptions for user %s\n", user.Username)
		es.DropAllSubscriptions(user.Token)
		es.InitSubscriptions(user.Token)
	}
}

func (es *EventSub) subscribeToEvent(userToken string, request twitch.SubscriptionRequest) error {
	twitchUrl := "https://api.twitch.tv/helix/eventsub/subscriptions"

	httpClient := &http.Client{}
	bytes, _ := json.Marshal(request)
	req, err := http.NewRequest("POST", twitchUrl, strings.NewReader(string(bytes)))
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", "Bearer "+userToken)
	req.Header.Add("Client-Id", os.Getenv("TWITCH_CLIENT_ID"))
	req.Header.Add("Content-Type", "application/json")

	res, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return err
	}

	subcriptionResponse := &twitch.SubscriptionResponse{}
	err = json.Unmarshal(body, subcriptionResponse)
	if err != nil {
		fmt.Println("Error unmarshalling response:", err)
		return err
	}
	if len(subcriptionResponse.Data) > 0 {
		if !(subcriptionResponse.Data[0].Status == "enabled") {
			return fmt.Errorf("subscription failed with status: %s\nResponse body : %s", subcriptionResponse.Data[0].Status, string(body))
		} else {
			return nil
		}
	} else {
		return fmt.Errorf("subscription failed with no data.\nResponse body: %s", string(body))
	}
}

func (es *EventSub) subscribeToMessageEvents(userToken string) error {
	broadcasterId, err := es.apiWrapper.GetUserInfoFromToken(userToken)
	if err != nil {
		return err
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

	err = es.subscribeToEvent(userToken, data)
	if err != nil {
		return err
	}
	return nil
}

func (es *EventSub) subscribeToRedemptionEvents(userToken string) error {
	broadcasterId, err := es.apiWrapper.GetUserInfoFromToken(userToken)
	if err != nil {
		return err
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

	err = es.subscribeToEvent(userToken, data)
	if err != nil {
		return err
	}
	return nil
}

func (es *EventSub) subscribeToPollEvents(userToken string) error {
	broadcasterId, err := es.apiWrapper.GetUserInfoFromToken(userToken)
	if err != nil {
		return err
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

	err = es.subscribeToEvent(userToken, data)
	if err != nil {
		return err
	}
	return nil
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
