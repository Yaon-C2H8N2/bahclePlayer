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

	notifcationHandler := GetNotificationHandler(pm.apiWrapper, token, conn)

	eventSub := pm.eventSubs[token]
	unsubscribe := eventSub.OnEvent(func(event twitch.NotificationMessage) {
		eventBytes, _ := json.Marshal(event)
		notifcationHandler.Handle(eventBytes)
	})

	for {
		err := conn.WriteMessage(websocket.PingMessage, nil)
		if err != nil {
			unsubscribe()
			conn.Close()

			pm.mutex.Lock()
			delete(pm.clients, token)
			pm.mutex.Unlock()

			break
		}
		time.Sleep(30 * time.Second)
	}
}
