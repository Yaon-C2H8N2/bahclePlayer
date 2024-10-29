package models

import (
	"fmt"
	"github.com/Yaon-C2H8N2/bahclePlayer/models/twitch"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
)

type PlayersManager struct {
	mutex    *sync.Mutex
	clients  map[string]*websocket.Conn
	upgrader websocket.Upgrader
}

func DefaultPlayersManager() *PlayersManager {
	return &PlayersManager{
		mutex:   &sync.Mutex{},
		clients: make(map[string]*websocket.Conn),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func (pm *PlayersManager) CreatePlayer(c *gin.Context) {
	pm.mutex.Lock()
	if _, ok := pm.clients[c.Query("access_token")]; !ok {
		conn, err := pm.upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			c.JSON(500, gin.H{
				"message": "Failed to upgrade connection",
				"error":   err.Error(),
			})
			return
		}

		token := c.Query("access_token")
		pm.clients[token] = conn

		go pm.mainLoop(token)
	} else {
		c.JSON(400, gin.H{
			"message": "Player already exists",
		})
	}
	pm.mutex.Unlock()
}

func (pm *PlayersManager) mainLoop(token string) {
	conn := pm.clients[token]

	type testMessage struct {
		MessageId string `json:"message_id"`
		Message   string `json:"message"`
	}

	eventSub := twitch.GetEventSub(token)
	eventSub.OnEvent(func(event any) {
		err := conn.WriteJSON(testMessage{MessageId: uuid.NewString(), Message: event.(string)})
		fmt.Printf("Sent message to client %s\n", token)
		if err != nil {
			conn.Close()
			pm.mutex.Lock()
			delete(pm.clients, token)
			pm.mutex.Unlock()
		}
	})
	eventSub.OnStarted(func() {
		eventSub.SubscribeToMessageEvents()
	})
	eventSub.Start()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Client disconnected")
			conn.Close()

			pm.mutex.Lock()
			delete(pm.clients, token)
			pm.mutex.Unlock()

			break
		}
	}
}
