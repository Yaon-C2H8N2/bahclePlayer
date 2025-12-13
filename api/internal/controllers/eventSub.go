package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/Yaon-C2H8N2/bahclePlayer/internal/models"
	"github.com/Yaon-C2H8N2/bahclePlayer/internal/models/twitch"
	"github.com/gorilla/websocket"
)

type EventListenerKeys int

const (
	EventListenerOnStarted EventListenerKeys = iota
	EventListenerOnStopped
	EventListenerOnError
	EventListenerOnRefresh
)

type EventSub struct {
	sessionId           string
	apiWrapper          *ApiWrapper
	notificationHandler *NotificationHandler
	user                models.Users
	twitchUser          twitch.UserInfo
	webSocketUrl        string
	lastEventListenerId int
	eventListeners      map[EventListenerKeys]map[int]func(*EventSub, any)
	eventListenersMutex sync.Mutex
	mainCtxCancel       context.CancelFunc
	loopCtxCancel       context.CancelFunc
}

// GetEventSub creates a new twitch EventSub instance. It uses the WebSocket protocol for receiving events.
// Returned errors are only for initialization errors.
// Further error handling must be handled via event listeners. See EventSub.AddEventListener for more details.
// Call EventSub.Start() to start the EventSub instance.
func GetEventSub(apiWrapper *ApiWrapper, user models.Users, webSocketUrl string) (*EventSub, error) {
	twitchUser, err := apiWrapper.GetUserInfoFromToken(user.Token)

	if err != nil {
		return nil, err
	}

	var newEventSub = &EventSub{
		user:                user,
		twitchUser:          twitchUser,
		apiWrapper:          apiWrapper,
		webSocketUrl:        webSocketUrl,
		lastEventListenerId: 0,
		eventListeners:      make(map[EventListenerKeys]map[int]func(*EventSub, any)),
		eventListenersMutex: sync.Mutex{},
	}

	newEventSub.notificationHandler = GetNotificationHandler(apiWrapper, user.Token)
	return newEventSub, nil
}

// Helper function to read the initial infos of a message from the websocket connection
func readMessageFromWebSocket(conn *websocket.Conn) (*twitch.BaseMessage, []byte, error) {
	var formattedErr error
	if conn == nil {
		formattedErr = fmt.Errorf("websocket connection is nil")
		return nil, nil, formattedErr
	}

	_, messageBytes, err := conn.ReadMessage()
	if err != nil {
		if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
			formattedErr = fmt.Errorf("unexpected websocket closure: %v", err)
		} else {
			formattedErr = fmt.Errorf("couldn't read message: %v", err)
		}
		return nil, nil, formattedErr
	}

	var message = &twitch.BaseMessage{}
	err = json.Unmarshal(messageBytes, message)
	if err != nil {
		formattedErr = fmt.Errorf("error unmarshalling base message: %v, raw data: %s", err, string(messageBytes))
		return nil, messageBytes, formattedErr
	}

	return message, messageBytes, nil
}

// Helper function to get all current EventSub subscriptions for a Twitch user from the Twitch API
func getAllSubscriptionsForTwitchUser(user models.Users) (twitch.SubscriptionResponse, error) {
	twitchUrl := os.Getenv("TWITCH_EVENTSUB_URL")
	httpClient := &http.Client{}

	req, err := http.NewRequest("GET", twitchUrl+"?user_id="+user.TwitchId, nil)
	if err != nil {
		return twitch.SubscriptionResponse{}, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Add("Authorization", "Bearer "+user.Token)
	req.Header.Add("Client-Id", os.Getenv("TWITCH_CLIENT_ID"))
	req.Header.Add("Content-Type", "application/json")

	res, err := httpClient.Do(req)
	if err != nil {
		return twitch.SubscriptionResponse{}, fmt.Errorf("error making request: %v", err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)

	if err != nil {
		return twitch.SubscriptionResponse{}, fmt.Errorf("error reading response: %v", err)
	}

	subscriptionResponse := &twitch.SubscriptionResponse{}
	err = json.Unmarshal(body, subscriptionResponse)
	if err != nil {
		return twitch.SubscriptionResponse{}, fmt.Errorf("error unmarshalling response: %v, raw data: %s", err, string(body))
	}
	return *subscriptionResponse, nil
}

// Start listens for Twitch EventSub events. It must be called after creating the EventSub instance with GetEventSub.
// Errors during operation are dispatched to the EventListenerOnError event listeners.
// To stop listening for events, call EventSub.Stop().
func (es *EventSub) Start() {
	unsub := es.AddEventListener(EventListenerOnStarted, func(es *EventSub, data any) {
		err := es.dropAllSubscriptions()
		if err != nil {
			es.dispatchToEventListeners(EventListenerOnError, fmt.Errorf("error dropping subscriptions: %v", err))
			return
		}
		err = es.initSubscriptions()
		if err != nil {
			es.dispatchToEventListeners(EventListenerOnError, fmt.Errorf("error initializing subscriptions: %v", err))
			return
		}
	})

	messageChan := make(chan chanContent)

	mainCtx, mainCtxCancel := context.WithCancel(context.Background())
	context.AfterFunc(mainCtx, func() {
		unsub()
		close(messageChan)
	})
	es.mainCtxCancel = mainCtxCancel

	es.startLoops(mainCtx, messageChan)
}

// Stop forces the EventSub instance to stop listening for events.
// Once stopped, the EventSub instance shouldn't be restarted. A new instance should be created instead.
func (es *EventSub) Stop() {
	if es.mainCtxCancel != nil {
		es.mainCtxCancel()
	}
}

// AddEventListener adds an event listener for the specified event key.
// The listener function will be called with the EventSub instance and event data when the event occurs.
// It returns a cancel function that can be called to remove the event listener.
// TODO : further document which events send what data.
func (es *EventSub) AddEventListener(key EventListenerKeys, listener func(*EventSub, any)) func() {
	es.eventListenersMutex.Lock()
	defer es.eventListenersMutex.Unlock()

	if es.eventListeners[key] == nil {
		es.eventListeners[key] = make(map[int]func(*EventSub, any))
	}

	listenerId := es.lastEventListenerId + 1
	es.lastEventListenerId = listenerId

	es.eventListeners[key][listenerId] = listener

	return func() {
		es.eventListenersMutex.Lock()
		defer es.eventListenersMutex.Unlock()

		delete(es.eventListeners[key], listenerId)
	}
}

// SetUser updates the EventSub instance to use a new user.
// This is useful when the user's token has been refreshed.
// It returns an error if the user information cannot be retrieved.
func (es *EventSub) SetUser(user models.Users) error {
	es.user = user

	twitchUser, err := es.apiWrapper.GetUserInfoFromToken(es.user.Token)
	if err != nil {
		return err
	}
	es.twitchUser = twitchUser

	return nil
}

// Internal function to start the read and process loops. Should not be called directly.
func (es *EventSub) startLoops(ctx context.Context, messageChan chan chanContent) {
	loopCtx, loopCtxCancel := context.WithCancel(ctx)
	context.AfterFunc(loopCtx, func() {
		select {
		case <-ctx.Done():
			return
		default:
			es.startLoops(ctx, messageChan)
		}
	})

	es.loopCtxCancel = loopCtxCancel

	go es.readLoop(loopCtx, messageChan)
	go es.processLoop(loopCtx, messageChan)
}

// Internal function to dispatch events to registered event listeners. Non-blocking.
func (es *EventSub) dispatchToEventListeners(key EventListenerKeys, data any) {
	go func() {
		es.eventListenersMutex.Lock()
		defer es.eventListenersMutex.Unlock()

		eventListenersCopy := make(map[int]func(*EventSub, any))
		for id, listener := range es.eventListeners[key] {
			eventListenersCopy[id] = listener
		}

		for _, listener := range eventListenersCopy {
			go listener(es, data)
		}
	}()
}

// Internal function to drop all current EventSub subscriptions for the Twitch user.
func (es *EventSub) dropAllSubscriptions() error {
	subscriptionResponse, err := getAllSubscriptionsForTwitchUser(es.user)
	if err != nil {
		return fmt.Errorf("error getting all subscriptions: %v", err)
	}

	var unsubscribeErrors []error
	for _, subscription := range subscriptionResponse.Data {
		if subscription.Status == "enabled" {
			err = es.unsubscribeFromEvent(subscription.ID)
			if err != nil {
				unsubscribeErrors = append(unsubscribeErrors, fmt.Errorf("error unsubscribing from %s: %v", subscription.Type, err))
			}
		}
	}

	if len(unsubscribeErrors) > 0 {
		return fmt.Errorf("errors occurred while unsubscribing: %v", unsubscribeErrors)
	}
	return nil
}

// Internal function to initialize EventSub subscriptions for the Twitch user.
func (es *EventSub) initSubscriptions() error {
	var subscriptionMethods = []func() error{
		es.subscribeToMessageEvents,
		es.subscribeToRedemptionEvents,
		es.subscribeToPollEvents,
	}

	for _, method := range subscriptionMethods {
		err := method()
		if err != nil {
			return err
		}
	}

	return nil
}

// Internal function to unsubscribe from a specific EventSub subscription by ID.
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
		errMsg := fmt.Errorf("error reading response: %v", err)
		return errMsg
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
		return fmt.Errorf("error reading response: %v", err)
	}

	subcriptionResponse := &twitch.SubscriptionResponse{}
	err = json.Unmarshal(body, subcriptionResponse)
	if err != nil {
		return fmt.Errorf("error unmarshalling response: %v, raw data: %s", err, string(body))
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

// Internal function to subscribe to channel chat message events.
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

// Internal function to subscribe to channel points redemption events.
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

// Internal function to subscribe to channel poll end events.
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

// Internal struct to pass data between read and process loops.
type chanContent struct {
	Message      *twitch.BaseMessage
	MessageBytes []byte
	error        error
}

// Main readLoop to read messages from the Twitch WebSocket connection.
// Messages are sent to the provided messageChan for processing.
// The loop runs until the provided context is cancelled.
// Should not be called directly.
func (es *EventSub) readLoop(ctx context.Context, messageChan chan chanContent) {
	if es.webSocketUrl == "" {
		errMsg := fmt.Errorf("websocket URL is empty")
		es.dispatchToEventListeners(EventListenerOnError, errMsg)
		return
	}

	parsedURL, err := url.Parse(es.webSocketUrl)
	if err != nil {
		errMsg := fmt.Errorf("invalid websocket URL format: %s, error: %v", es.webSocketUrl, err)
		es.dispatchToEventListeners(EventListenerOnError, errMsg)
		return
	}

	if parsedURL.Scheme != "ws" && parsedURL.Scheme != "wss" {
		errMsg := fmt.Errorf("invalid websocket URL scheme: %s (must be ws or wss)", parsedURL.Scheme)
		es.dispatchToEventListeners(EventListenerOnError, errMsg)
		return
	}

	conn, _, err := websocket.DefaultDialer.Dial(es.webSocketUrl, nil)
	defer conn.Close()

	if err != nil {
		errMsg := fmt.Errorf("couldn't dial twitch websocket: %s", err)
		es.dispatchToEventListeners(EventListenerOnError, errMsg)
		return
	}

	for {
		message, messageBytes, err := readMessageFromWebSocket(conn)

		select {
		case <-ctx.Done():
			return
		case messageChan <- chanContent{
			Message:      message,
			MessageBytes: messageBytes,
			error:        err,
		}:
		}

		if err != nil {
			es.dispatchToEventListeners(EventListenerOnError, err)
			return
		}
	}
}

// Main processLoop to handle messages received from the Twitch WebSocket connection.
// Messages are read from the provided messageChan.
// The loop runs until the provided context is cancelled.
// Should not be called directly.
func (es *EventSub) processLoop(ctx context.Context, messageChan chan chanContent) {
	defer es.Stop()

loopiloop:
	for {
		var content chanContent
		var ok bool

		select {
		case <-ctx.Done():
			es.dispatchToEventListeners(EventListenerOnStopped, nil)
			break loopiloop
		case content, ok = <-messageChan:
			break
		}

		if !ok || content.error != nil {
			var errMsg error
			if content.error != nil {
				errMsg = fmt.Errorf("error reading message: %v", content.error)
			} else {
				errMsg = fmt.Errorf("message channel closed unexpectedly")
			}

			es.dispatchToEventListeners(EventListenerOnError, errMsg)
			break loopiloop
		}

		message := content.Message
		messageBytes := content.MessageBytes

		switch message.Metadata.MessageType {
		case "session_welcome":
			var welcomeMessage = &twitch.WelcomeMessage{}
			err := json.Unmarshal(messageBytes, welcomeMessage)
			if err != nil {
				errMsg := fmt.Errorf("error unmarshalling welcome message: %v, raw data: %s", err, string(messageBytes))
				es.dispatchToEventListeners(EventListenerOnError, errMsg)
				break loopiloop
			}

			es.sessionId = welcomeMessage.Payload.Session.Id
			es.dispatchToEventListeners(EventListenerOnStarted, nil)
			break
		case "notification":
			var notificationMessage = &twitch.NotificationMessage{}
			err := json.Unmarshal(messageBytes, notificationMessage)
			if err != nil {
				errMsg := fmt.Errorf("error unmarshalling notification message: %v, raw data: %s", err, string(messageBytes))
				es.dispatchToEventListeners(EventListenerOnError, errMsg)
				break loopiloop
			}
			go es.notificationHandler.Handle(messageBytes)
			break
		case "session_reconnect":
			// The session_reconnect has the same structure as the session_welcome message
			var reconnectMessage = &twitch.WelcomeMessage{}
			err := json.Unmarshal(messageBytes, reconnectMessage)
			if err != nil {
				errMsg := fmt.Errorf("error unmarshalling reconnect message: %v, raw data: %s", err, string(messageBytes))
				es.dispatchToEventListeners(EventListenerOnError, errMsg)
				break loopiloop
			}

			reconnectUrl := reconnectMessage.Payload.Session.ReconnectUrl
			if reconnectUrl != "" {
				if _, parseErr := url.Parse(reconnectUrl); parseErr != nil {
					errMsg := fmt.Errorf("invalid reconnect URL: %s, error: %v", reconnectUrl, parseErr)
					es.dispatchToEventListeners(EventListenerOnError, errMsg)
					break loopiloop
				}
			}

			es.dispatchToEventListeners(EventListenerOnRefresh, reconnectUrl)

			es.loopCtxCancel()
			break
		case "session_keepalive":
			break
		default:
			errMsg := fmt.Errorf("received unknown message type: %s", message.Metadata.MessageType)
			es.dispatchToEventListeners(EventListenerOnError, errMsg)
			break
		}
	}
	return
}
