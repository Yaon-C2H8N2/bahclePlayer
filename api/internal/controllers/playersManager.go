package controllers

import (
	"encoding/json"
	"github.com/Yaon-C2H8N2/bahclePlayer/internal/models"
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

func (pm *PlayersManager) GetConnFromToken(token string) *websocket.Conn {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	conn := pm.clients[token]

	return conn
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
		time.Sleep(10 * time.Second)
		if token == "" {
			payload := struct {
				Error   string `json:"error"`
				Message string `json:"message"`
			}{
				Error:   "Timeout",
				Message: "No message received before timeout",
			}

			conn.WriteJSON(payload)
			conn.Close()
		}
	}()

	pm.mutex.Lock()
	welcome := struct {
		Welcome string `json:"welcome"`
	}{
		Welcome: "Socket connection established",
	}
	conn.WriteJSON(welcome)

	_, messageBytes, err := conn.ReadMessage()
	if err != nil {
		c.JSON(500, gin.H{
			"message": "An error occured when receiving welcome message",
			"error":   err.Error(),
		})
		return
	}

	var message struct {
		Token string `json:"token"`
	}
	json.Unmarshal(messageBytes, &message)
	token = message.Token

	//TODO : implement token verification with twitch

	if token != "" {
		pm.clients[message.Token] = conn

		go pm.mainLoop(token)
	}
	pm.mutex.Unlock()
}

func (pm *PlayersManager) mainLoop(token string) {
	conn := pm.clients[token]

	eventSub := pm.eventSubs[token]
	unsubscribeEvent := eventSub.notificationHandler.OnEvent(func(newVideo models.UsersVideos) {
		err := conn.WriteJSON(newVideo)
		if err != nil {
			payload := struct {
				Error   string `json:"error"`
				Message string `json:"message"`
			}{
				Error:   err.Error(),
				Message: "An error occured when sending new video",
			}

			conn.WriteJSON(payload)
		}
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
		err := conn.WriteControl(websocket.PingMessage, []byte("PING!"), time.Now().Add(5*time.Second))
		if err != nil {
			unsubscribeEvent()
			unsubscribeError()
			conn.Close()

			pm.mutex.Lock()
			delete(pm.clients, token)
			pm.mutex.Unlock()

			break
		}
		time.Sleep(5 * time.Second)
	}
}
