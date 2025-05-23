package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/Yaon-C2H8N2/bahclePlayer/internal/models"
	"github.com/Yaon-C2H8N2/bahclePlayer/internal/models/twitch"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type EventSub struct {
	onStarted           func()
	onError             func(error error)
	sessionId           string
	apiWrapper          *ApiWrapper
	notificationHandler *NotificationHandler
	user                models.Users
	twitchUser          twitch.UserInfo
}

func GetForAllUsers(apiWrapper *ApiWrapper) map[string]*EventSub {
	users, err := models.GetAllUsers()
	if err != nil {
		fmt.Println("Error getting users:", err)
		return nil
	}

	eventSubs := make(map[string]*EventSub)

	for _, user := range users {
		if user.Token == "" {
			continue
		}
		fmt.Printf("Initializing subscriptions for user %s\n", user.Username)
		es, err := GetEventSub(apiWrapper, user)

		if err != nil {
			tokenResponse, err := RefreshUserToken(user.RefreshToken)
			if err != nil {
				fmt.Printf("Error refreshing token for user %s: %s\n", user.Username, err)
				continue
			}
			fmt.Printf("Refreshing token for user %s\n", user.Username)

			user, err = models.AddOrUpdateUser(user, *tokenResponse)
			if err != nil {
				fmt.Printf("Error updating user %s: %s\n", user.Username, err)
				continue
			}

			es, err = GetEventSub(apiWrapper, user)
			if err != nil {
				fmt.Printf("Error getting event sub for user %s: %s\n", user.Username, err)
				continue
			}
		}
		eventSubs[user.TwitchId] = es
	}

	return eventSubs
}

func GetEventSub(apiWrapper *ApiWrapper, user models.Users) (*EventSub, error) {
	twitchUser, err := apiWrapper.GetUserInfoFromToken(user.Token)

	if err != nil {
		fmt.Println("Error getting user info from token:", err)
		return nil, err
	}
	var newEventSub = &EventSub{
		user:       user,
		twitchUser: twitchUser,
	}

	newEventSub.apiWrapper = apiWrapper
	newEventSub.notificationHandler = GetNotificationHandler(apiWrapper, user.Token)
	return newEventSub, nil
}

func (es *EventSub) Start() {
	es.listenToMessages()
}

func (es *EventSub) UpdateUser(user models.Users) {
	// todo : implement a graceful stop before updating the user

	es.user = user
	es.listenToMessages()
}

func (es *EventSub) OnError(callback func(err error)) func() {
	es.onError = callback

	return func() {
		es.onError = nil
	}
}

func (es *EventSub) OnStarted(callback func()) func() {
	es.onStarted = callback

	return func() {
		es.onStarted = nil
	}
}

func (es *EventSub) GetAllSubscriptionsForTwitchUser() (twitch.SubscriptionResponse, error) {
	twitchUrl := os.Getenv("TWITCH_EVENTSUB_URL")
	httpClient := &http.Client{}

	req, err := http.NewRequest("GET", twitchUrl+"?user_id="+es.twitchUser.ID, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return twitch.SubscriptionResponse{}, err
	}

	req.Header.Add("Authorization", "Bearer "+es.user.Token)
	req.Header.Add("Client-Id", os.Getenv("TWITCH_CLIENT_ID"))
	req.Header.Add("Content-Type", "application/json")

	res, err := httpClient.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return twitch.SubscriptionResponse{}, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)

	if err != nil {
		fmt.Println("Error reading response:", err)
		return twitch.SubscriptionResponse{}, err
	}

	subscriptionResponse := &twitch.SubscriptionResponse{}
	err = json.Unmarshal(body, subscriptionResponse)
	if err != nil {
		fmt.Println("Error unmarshalling response:", err)
		return twitch.SubscriptionResponse{}, err
	}
	return *subscriptionResponse, nil
}

func (es *EventSub) DropAllSubscriptions() {
	subscriptionResponse, err := es.GetAllSubscriptionsForTwitchUser()
	if err != nil {
		fmt.Println("Error getting all subscriptions:", err)
		return
	}

	fmt.Printf("eventSub[%s] dropping %d subscriptions\n", es.user.Username, len(subscriptionResponse.Data))
	for _, subscription := range subscriptionResponse.Data {
		if subscription.Status == "enabled" {
			err = es.unsubscribeFromEvent(subscription.ID)
			if err != nil {
				fmt.Println("Error unsubscribing from event:", err)
			}
		}
	}
}

func (es *EventSub) unsubscribeFromEvent(subscriptionId string) error {
	twitchUrl := os.Getenv("TWITCH_EVENTSUB_URL")

	httpClient := &http.Client{}
	req, err := http.NewRequest("DELETE", twitchUrl+"?id="+subscriptionId, nil)
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", "Bearer "+es.user.Token)
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

	if res.StatusCode != 204 {
		return fmt.Errorf("failed to unsubscribe from event: %s\nResponse body: %s", res.Status, string(body))
	}
	return nil
}

func (es *EventSub) InitSubscriptions() {
	var err error
	err = es.subscribeToMessageEvents(es.user.Token)
	if err != nil {
		errorMessage := fmt.Sprintf("Error subscribing to message events with token %s: %s", es.user.Token, err)
		fmt.Printf(errorMessage)
		if es.onError != nil {
			go es.onError(fmt.Errorf(errorMessage))
		}
	}
	err = es.subscribeToRedemptionEvents()
	if err != nil {
		errorMessage := fmt.Sprintf("Error subscribing to redemption events with token %s: %s\n", es.user.Token, err)
		fmt.Printf(errorMessage)
		if es.onError != nil {
			go es.onError(fmt.Errorf(errorMessage))
		}
	}
	err = es.subscribeToPollEvents()
	if err != nil {
		errorMessage := fmt.Sprintf("Error subscribing to poll events with token %s: %s\n", es.user.Token, err)
		fmt.Printf(errorMessage)
		if es.onError != nil {
			go es.onError(fmt.Errorf(errorMessage))
		}
	}
}

func (es *EventSub) subscribeToEvent(request twitch.SubscriptionRequest) error {
	twitchUrl := os.Getenv("TWITCH_EVENTSUB_URL")

	httpClient := &http.Client{}
	bytes, _ := json.Marshal(request)
	req, err := http.NewRequest("POST", twitchUrl, strings.NewReader(string(bytes)))
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", "Bearer "+es.user.Token)
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
	var data = twitch.SubscriptionRequest{
		Type:    "channel.chat.message",
		Version: "1",
		Condition: twitch.Condition{
			BroadcasterUserId: es.twitchUser.ID,
			UserId:            es.twitchUser.ID,
		},
		Transport: twitch.Transport{
			Method:    "websocket",
			SessionId: es.sessionId,
		},
	}

	err := es.subscribeToEvent(data)
	if err != nil {
		return err
	}
	return nil
}

func (es *EventSub) subscribeToRedemptionEvents() error {
	broadcasterId, err := es.apiWrapper.GetUserInfoFromToken(es.user.Token)
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

	err = es.subscribeToEvent(data)
	if err != nil {
		return err
	}
	return nil
}

func (es *EventSub) subscribeToPollEvents() error {
	var data = twitch.SubscriptionRequest{
		Type:    "channel.poll.end",
		Version: "1",
		Condition: twitch.Condition{
			BroadcasterUserId: es.twitchUser.ID,
		},
		Transport: twitch.Transport{
			Method:    "websocket",
			SessionId: es.sessionId,
		},
	}

	err := es.subscribeToEvent(data)
	if err != nil {
		return err
	}
	return nil
}

func (es *EventSub) listenToMessages() {
	//https://github.com/gorilla/websocket/blob/main/examples/echo/client.go
	webSocketUrl := os.Getenv("TWITCH_EVENTSUB_WEBSOCKET_URL")
	conn, _, err := websocket.DefaultDialer.Dial(webSocketUrl, nil)

	if err != nil {
		errMsg := fmt.Sprintf("eventSub[%s] couldn't dial twitch websocket: %s", es.user.Username, err)
		log.Println(errMsg)
		if es.onError != nil {
			go es.onError(fmt.Errorf(errMsg))
		}
	}

	go func() {
		defer conn.Close()
	loopiloop:
		for {
			_, messageBytes, err := conn.ReadMessage()
			if err != nil {
				errMsg := fmt.Sprintf("eventSub[%s] couldn't read message: %s", es.user.Username, messageBytes)
				log.Println(errMsg)
				if es.onError != nil {
					go es.onError(fmt.Errorf(errMsg))
				}
				break loopiloop
			}

			var message = &twitch.BaseMessage{}
			err = json.Unmarshal(messageBytes, message)
			if err != nil {
				errMsg := fmt.Sprintf("eventSub[%s] error unmarshalling base message: %s", es.user.Username, messageBytes)
				log.Println(errMsg)
				if es.onError != nil {
					go es.onError(fmt.Errorf(errMsg))
				}
				break loopiloop
			}

			switch message.Metadata.MessageType {
			case "session_welcome":
				var welcomeMessage = &twitch.WelcomeMessage{}
				err = json.Unmarshal(messageBytes, welcomeMessage)
				if err != nil {
					errMsg := fmt.Sprintf("eventSub[%s] error unmarshalling welcome message: %s", es.user.Username, messageBytes)
					log.Println(errMsg)
					if es.onError != nil {
						go es.onError(fmt.Errorf(errMsg))
					}
					break loopiloop
				}

				es.sessionId = welcomeMessage.Payload.Session.Id
				go es.onStarted()
				break
			case "notification":
				var notificationMessage = &twitch.NotificationMessage{}
				err = json.Unmarshal(messageBytes, notificationMessage)
				if err != nil {
					errMsg := fmt.Sprintf("eventSub[%s] error unmarshalling notification message: %s", es.user.Username, messageBytes)
					log.Println(errMsg)
					if es.onError != nil {
						go es.onError(fmt.Errorf(errMsg))
					}
					break loopiloop
				}
				fmt.Printf("eventSub[%s] received notification: %s\n", es.user.Username, notificationMessage.Metadata.MessageType)
				go es.notificationHandler.Handle(messageBytes)
				break
			}
		}
		fmt.Printf("eventSub[%s] stopped %s\n", es.user.Username, err)
	}()
}
