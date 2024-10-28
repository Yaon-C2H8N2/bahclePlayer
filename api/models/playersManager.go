package models

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"math/rand"
	"net/http"
	"sync"
	"time"
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
		go pm.testNotify()
	} else {
		c.JSON(400, gin.H{
			"message": "Player already exists",
		})
	}
	pm.mutex.Unlock()
}

func (pm *PlayersManager) mainLoop(token string) {
	conn := pm.clients[token]

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

type testMessage struct {
	MessageId string `json:"message_id"`
	Message   string `json:"message"`
}

func (pm *PlayersManager) testNotify() {
	for {
		for key, conn := range pm.clients {
			err := conn.WriteJSON(testMessage{MessageId: uuid.NewString(), Message: "Hello"})
			fmt.Printf("Sent message to client %s\n", key)
			if err != nil {
				conn.Close()
				pm.mutex.Lock()
				delete(pm.clients, key)
				pm.mutex.Unlock()
			}
		}
		time.Sleep(time.Duration(rand.Intn(5-1)+1) * time.Second)
	}
}
