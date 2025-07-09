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
	onStarted           func(es *EventSub)
	onError             func(es *EventSub, error error)
	onRefresh           func(es *EventSub, url string)
	sessionId           string
	apiWrapper          *ApiWrapper
	notificationHandler *NotificationHandler
	user                models.Users
	twitchUser          twitch.UserInfo
	webSocketUrl        string
	stopChan            chan struct{}
}

func GetEventSub(apiWrapper *ApiWrapper, user models.Users, webSocketUrl string) (*EventSub, error) {
	twitchUser, err := apiWrapper.GetUserInfoFromToken(user.Token)

	if err != nil {
		fmt.Println("Error getting user info from token:", err)
		return nil, err
	}
	var newEventSub = &EventSub{
		user:         user,
		twitchUser:   twitchUser,
		apiWrapper:   apiWrapper,
		webSocketUrl: webSocketUrl,
	}

	newEventSub.notificationHandler = GetNotificationHandler(apiWrapper, user.Token)
	return newEventSub, nil
}

func (es *EventSub) Start() {
	es.listenToMessages()
	es.stopChan = make(chan struct{})
}

func (es *EventSub) Stop() {
	close(es.stopChan)
}

func (es *EventSub) OnError(callback func(es *EventSub, err error)) func() {
	es.onError = callback

	return func() {
		es.onError = nil
	}
}

func (es *EventSub) OnStarted(callback func(es *EventSub)) func() {
	es.onStarted = callback

	return func() {
		es.onStarted = nil
	}
}

func (es *EventSub) OnRefresh(callback func(es *EventSub, url string)) func() {
	es.onRefresh = callback
	return func() {
		es.onRefresh = nil
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

func (es *EventSub) InitSubscriptions() {
	var err error
	err = es.subscribeToMessageEvents()
	if err != nil {
		errorMessage := fmt.Sprintf("Error subscribing to message events with token %s: %s", es.user.Token, err)
		fmt.Printf(errorMessage)
		if es.onError != nil {
			go es.onError(es, fmt.Errorf(errorMessage))
		}
	}
	err = es.subscribeToRedemptionEvents()
	if err != nil {
		errorMessage := fmt.Sprintf("Error subscribing to redemption events with token %s: %s\n", es.user.Token, err)
		fmt.Printf(errorMessage)
		if es.onError != nil {
			go es.onError(es, fmt.Errorf(errorMessage))
		}
	}
	err = es.subscribeToPollEvents()
	if err != nil {
		errorMessage := fmt.Sprintf("Error subscribing to poll events with token %s: %s\n", es.user.Token, err)
		fmt.Printf(errorMessage)
		if es.onError != nil {
			go es.onError(es, fmt.Errorf(errorMessage))
		}
	}
}

func (es *EventSub) SetUser(user models.Users) error {
	es.user = user

	twitchUser, err := es.apiWrapper.GetUserInfoFromToken(es.user.Token)
	if err != nil {
		return err
	}
	es.twitchUser = twitchUser

	return nil
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

func (es *EventSub) subscribeToMessageEvents() error {
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

func (es *EventSub) readMessageFromWebSocket(conn *websocket.Conn) (*twitch.BaseMessage, []byte, error) {
	_, messageBytes, err := conn.ReadMessage()
	if err != nil {
		errMsg := fmt.Sprintf("eventSub[%s] couldn't read message", es.user.Username)
		log.Println(errMsg)
		if es.onError != nil {
			go es.onError(es, fmt.Errorf(errMsg))
		}
		return nil, nil, err
	}

	var message = &twitch.BaseMessage{}
	err = json.Unmarshal(messageBytes, message)
	if err != nil {
		errMsg := fmt.Sprintf("eventSub[%s] error unmarshalling base message: %s", es.user.Username, messageBytes)
		log.Println(errMsg)
		if es.onError != nil {
			go es.onError(es, fmt.Errorf(errMsg))
		}
		return nil, messageBytes, err
	}

	return message, messageBytes, nil
}

type chanContent struct {
	Message      *twitch.BaseMessage
	MessageBytes []byte
	error        error
}

func (es *EventSub) getMessageChannel(conn *websocket.Conn) chan chanContent {
	messageChan := make(chan chanContent)

	go func() {
		defer close(messageChan)
		for {
			message, messageBytes, err := es.readMessageFromWebSocket(conn)

			messageChan <- chanContent{
				Message:      message,
				MessageBytes: messageBytes,
				error:        err,
			}

			if err != nil {
				return
			}
		}
	}()

	return messageChan
}

func (es *EventSub) listenToMessages() {
	fmt.Printf("eventSub[%s] starting message listener\n", es.user.Username)
	//https://github.com/gorilla/websocket/blob/main/examples/echo/client.go
	conn, _, err := websocket.DefaultDialer.Dial(es.webSocketUrl, nil)

	if err != nil {
		errMsg := fmt.Sprintf("eventSub[%s] couldn't dial twitch websocket: %s", es.user.Username, err)
		log.Println(errMsg)
		if es.onError != nil {
			go es.onError(es, fmt.Errorf(errMsg))
		}
		return
	}

	go func() {
		defer conn.Close()

		messageChan := es.getMessageChannel(conn)

	loopiloop:
		for {
			var content chanContent
			var ok bool

			select {
			case <-es.stopChan:
				fmt.Printf("eventSub[%s] stopping message listener\n", es.user.Username)
				break loopiloop
			case content, ok = <-messageChan:
				break
			}

			if !ok || content.error != nil {
				var errMsg string
				if content.error != nil {
					errMsg = fmt.Sprintf("eventSub[%s] error reading message: %v", es.user.Username, content.error)
				} else {
					errMsg = fmt.Sprintf("eventSub[%s] message channel closed", es.user.Username)
				}
				log.Println(errMsg)
				if es.onError != nil {
					go es.onError(es, fmt.Errorf(errMsg))
				}
				if es.onRefresh != nil {
					go es.onRefresh(es, os.Getenv("TWITCH_EVENTSUB_URL"))
				}
				continue
			}

			message := content.Message
			messageBytes := content.MessageBytes

			switch message.Metadata.MessageType {
			case "session_welcome":
				var welcomeMessage = &twitch.WelcomeMessage{}
				err = json.Unmarshal(messageBytes, welcomeMessage)
				if err != nil {
					errMsg := fmt.Sprintf("eventSub[%s] error unmarshalling welcome message: %s", es.user.Username, messageBytes)
					log.Println(errMsg)
					if es.onError != nil {
						go es.onError(es, fmt.Errorf(errMsg))
					}
					break loopiloop
				}

				es.sessionId = welcomeMessage.Payload.Session.Id
				if es.onStarted != nil {
					go es.onStarted(es)
				}
				break
			case "notification":
				var notificationMessage = &twitch.NotificationMessage{}
				err = json.Unmarshal(messageBytes, notificationMessage)
				if err != nil {
					errMsg := fmt.Sprintf("eventSub[%s] error unmarshalling notification message: %s", es.user.Username, messageBytes)
					log.Println(errMsg)
					if es.onError != nil {
						go es.onError(es, fmt.Errorf(errMsg))
					}
					break loopiloop
				}
				fmt.Printf("eventSub[%s] received notification: %s\n", es.user.Username, notificationMessage.Metadata.MessageType)
				go es.notificationHandler.Handle(messageBytes)
				break
			case "session_reconnect":
				// The session_reconnect has the same structure as the session_welcome message
				var reconnectMessage = &twitch.WelcomeMessage{}
				err = json.Unmarshal(messageBytes, reconnectMessage)
				if err != nil {
					errMsg := fmt.Sprintf("eventSub[%s] error unmarshalling reconnect message: %s", es.user.Username, messageBytes)
					log.Println(errMsg)
					if es.onError != nil {
						go es.onError(es, fmt.Errorf(errMsg))
					}
					break loopiloop
				}
				if es.onRefresh != nil {
					go es.onRefresh(es, reconnectMessage.Payload.Session.ReconnectUrl)
				}
				fmt.Printf("eventSub[%s] received session_reconnect, reconnecting to %s\n", es.user.Username, reconnectMessage.Payload.Session.ReconnectUrl)
				break
			case "session_keepalive":
				// This is a keepalive message, we can ignore it
				break
			default:
				errMsg := fmt.Sprintf("eventSub[%s] received unknown message type: %s", es.user.Username, message.Metadata.MessageType)
				log.Println(errMsg)
				if es.onError != nil {
					go es.onError(es, fmt.Errorf(errMsg))
				}
				break
			}
		}
		fmt.Printf("eventSub[%s] stopped\n", es.user.Username)
	}()
}
