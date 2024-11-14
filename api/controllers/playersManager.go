package controllers

import (
	"encoding/json"
	"github.com/Yaon-C2H8N2/bahclePlayer/models/twitch"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
	"time"
)

type PlayersManager struct {
	mutex      *sync.Mutex
	clients    map[string]*websocket.Conn
	upgrader   websocket.Upgrader
	eventSubs  map[string]*EventSub
	apiWrapper *ApiWrapper
}

func DefaultPlayersManager(eventSubs map[string]*EventSub, apiWrapper *ApiWrapper) *PlayersManager {
	return &PlayersManager{
		mutex:   &sync.Mutex{},
		clients: make(map[string]*websocket.Conn),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		eventSubs:  eventSubs,
		apiWrapper: apiWrapper,
	}
}

func (pm *PlayersManager) CreatePlayer(c *gin.Context) {
	conn, err := pm.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(500, gin.H{
			"message": "Failed to upgrade connection",
			"error":   err.Error(),
		})
		return
	}

	token := ""
	go func() {
		_, messageBytes, err := conn.ReadMessage()
		if err != nil {
			c.JSON(500, gin.H{
				"message": "An error occured when receiving welcome message",
				"error":   err.Error(),
			})
			return
		}

		var message struct{ token string }
		json.Unmarshal(messageBytes, &message)
		token = message.token

		//TODO : implement token verification with twitch

		pm.mutex.Lock()
		if _, ok := pm.clients[token]; !ok && token != "" {
			pm.clients[message.token] = conn

			go pm.mainLoop(token)
		} else {
			c.JSON(400, gin.H{
				"message": "Player already exists",
			})
		}
		pm.mutex.Unlock()
	}()

	time.Sleep(10 * time.Second)
	if token == "" {
		conn.Close()
	}
}

func (pm *PlayersManager) mainLoop(token string) {
	conn := pm.clients[token]

	notifcationHandler := GetNotificationHandler(pm.apiWrapper, token, conn)

	eventSub := pm.eventSubs[token]
	unsubscribeEvent := eventSub.OnEvent(func(event twitch.NotificationMessage) {
		eventBytes, _ := json.Marshal(event)
		notifcationHandler.Handle(eventBytes)
	})
	unsubscribeError := eventSub.OnError(func(eventSubError error) {
		payload := struct {
			Error   string `json:"error"`
			Message string `json:"message"`
		}{
			Error:   eventSubError.Error(),
			Message: "An error occured with Twitch event listener",
		}

		conn.WriteJSON(payload)
	})

	for {
		err := conn.WriteMessage(websocket.PingMessage, nil)
		if err != nil {
			unsubscribeEvent()
			unsubscribeError()
			conn.Close()

			pm.mutex.Lock()
			delete(pm.clients, token)
			pm.mutex.Unlock()

			break
		}
		time.Sleep(30 * time.Second)
	}
}
