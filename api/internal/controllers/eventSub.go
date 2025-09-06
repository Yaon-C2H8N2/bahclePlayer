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
	"net/url"
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
	conn                *websocket.Conn
	isConnected         bool
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
	es.stopChan = make(chan struct{})
	es.isConnected = false
	es.listenToMessages()
}

func (es *EventSub) Stop() {
	es.isConnected = false
	if es.conn != nil {
		es.conn.Close()
		es.conn = nil
	}
	if es.stopChan != nil {
		close(es.stopChan)
		es.stopChan = nil
	}
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
	if conn == nil {
		err := fmt.Errorf("websocket connection is nil")
		errMsg := fmt.Sprintf("eventSub[%s] websocket connection is nil", es.user.Username)
		log.Println(errMsg)
		if es.onError != nil {
			go es.onError(es, err)
		}
		return nil, nil, err
	}

	_, messageBytes, err := conn.ReadMessage()
	if err != nil {
		// More detailed error logging
		if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
			errMsg := fmt.Sprintf("eventSub[%s] unexpected websocket close: %v", es.user.Username, err)
			log.Println(errMsg)
		} else {
			errMsg := fmt.Sprintf("eventSub[%s] couldn't read message: %v", es.user.Username, err)
			log.Println(errMsg)
		}
		if es.onError != nil {
			go es.onError(es, err)
		}
		return nil, nil, err
	}

	var message = &twitch.BaseMessage{}
	err = json.Unmarshal(messageBytes, message)
	if err != nil {
		errMsg := fmt.Sprintf("eventSub[%s] error unmarshalling base message: %v, raw data: %s", es.user.Username, err, string(messageBytes))
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
			select {
			case <-es.stopChan:
				return
			default:
			}

			if conn == nil || !es.isConnected {
				messageChan <- chanContent{
					Message:      nil,
					MessageBytes: nil,
					error:        fmt.Errorf("websocket connection is not available"),
				}
				return
			}

			message, messageBytes, err := es.readMessageFromWebSocket(conn)

			messageChan <- chanContent{
				Message:      message,
				MessageBytes: messageBytes,
				error:        err,
			}

			if err != nil {
				es.isConnected = false
				return
			}
		}
	}()

	return messageChan
}

func (es *EventSub) listenToMessages() {
	fmt.Printf("eventSub[%s] starting message listener\n", es.user.Username)

	if es.webSocketUrl == "" {
		errMsg := fmt.Sprintf("eventSub[%s] websocket URL is empty", es.user.Username)
		log.Println(errMsg)
		if es.onError != nil {
			go es.onError(es, fmt.Errorf(errMsg))
		}
		return
	}

	parsedURL, err := url.Parse(es.webSocketUrl)
	if err != nil {
		errMsg := fmt.Sprintf("eventSub[%s] invalid websocket URL format: %s, error: %v", es.user.Username, es.webSocketUrl, err)
		log.Println(errMsg)
		if es.onError != nil {
			go es.onError(es, fmt.Errorf(errMsg))
		}
		return
	}

	if parsedURL.Scheme != "ws" && parsedURL.Scheme != "wss" {
		errMsg := fmt.Sprintf("eventSub[%s] invalid websocket URL scheme: %s (must be ws or wss)", es.user.Username, parsedURL.Scheme)
		log.Println(errMsg)
		if es.onError != nil {
			go es.onError(es, fmt.Errorf(errMsg))
		}
		return
	}

	conn, _, err := websocket.DefaultDialer.Dial(es.webSocketUrl, nil)

	if err != nil {
		errMsg := fmt.Sprintf("eventSub[%s] couldn't dial twitch websocket: %s", es.user.Username, err)
		log.Println(errMsg)
		if es.onError != nil {
			go es.onError(es, fmt.Errorf(errMsg))
		}
		return
	}

	es.conn = conn
	es.isConnected = true

	go func() {
		defer func() {
			es.isConnected = false
			if es.conn != nil {
				es.conn.Close()
				es.conn = nil
			}
		}()

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
				es.isConnected = false

				select {
				case <-es.stopChan:
					fmt.Printf("eventSub[%s] stopping due to stop signal\n", es.user.Username)
				default:
					if es.onError != nil {
						go es.onError(es, fmt.Errorf(errMsg))
					}
					if es.onRefresh != nil && es.webSocketUrl != "" {
						go es.onRefresh(es, es.webSocketUrl)
					}
				}
				break loopiloop
			}

			message := content.Message
			messageBytes := content.MessageBytes

			switch message.Metadata.MessageType {
			case "session_welcome":
				var welcomeMessage = &twitch.WelcomeMessage{}
				err = json.Unmarshal(messageBytes, welcomeMessage)
				if err != nil {
					errMsg := fmt.Sprintf("eventSub[%s] error unmarshalling welcome message: %v, raw data: %s", es.user.Username, err, string(messageBytes))
					log.Println(errMsg)
					if es.onError != nil {
						go es.onError(es, fmt.Errorf(errMsg))
					}
					break loopiloop
				}

				es.sessionId = welcomeMessage.Payload.Session.Id
				fmt.Printf("eventSub[%s] received session_welcome, session_id: %s\n", es.user.Username, es.sessionId)
				if es.onStarted != nil {
					go es.onStarted(es)
				}
				break
			case "notification":
				var notificationMessage = &twitch.NotificationMessage{}
				err = json.Unmarshal(messageBytes, notificationMessage)
				if err != nil {
					errMsg := fmt.Sprintf("eventSub[%s] error unmarshalling notification message: %v, raw data: %s", es.user.Username, err, string(messageBytes))
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
					errMsg := fmt.Sprintf("eventSub[%s] error unmarshalling reconnect message: %v, raw data: %s", es.user.Username, err, string(messageBytes))
					log.Println(errMsg)
					if es.onError != nil {
						go es.onError(es, fmt.Errorf(errMsg))
					}
					break loopiloop
				}

				reconnectUrl := reconnectMessage.Payload.Session.ReconnectUrl
				if reconnectUrl != "" {
					if _, parseErr := url.Parse(reconnectUrl); parseErr != nil {
						errMsg := fmt.Sprintf("eventSub[%s] invalid reconnect URL: %s, error: %v", es.user.Username, reconnectUrl, parseErr)
						log.Println(errMsg)
						if es.onError != nil {
							go es.onError(es, fmt.Errorf(errMsg))
						}
						break loopiloop
					}
				}

				if es.onRefresh != nil {
					go es.onRefresh(es, reconnectUrl)
				}
				fmt.Printf("eventSub[%s] received session_reconnect, reconnecting to %s\n", es.user.Username, reconnectUrl)
				break
			case "session_keepalive":
				// This is a keepalive message, we can ignore it
				fmt.Printf("eventSub[%s] received keepalive\n", es.user.Username)
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
