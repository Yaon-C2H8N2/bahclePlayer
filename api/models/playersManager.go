package models

import (
	"fmt"
	"github.com/Yaon-C2H8N2/bahclePlayer/models/twitch"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
	"os"
	"sync"
)

type PlayersManager struct {
	mutex      *sync.Mutex
	clients    map[string]*websocket.Conn
	upgrader   websocket.Upgrader
	eventSub   *twitch.EventSub
	apiWrapper *twitch.ApiWrapper
}

func DefaultPlayersManager() *PlayersManager {
	apiWrapper := twitch.GetApiWrapper()
	appToken, err := twitch.RequestAppToken(os.Getenv("TWITCH_CLIENT_ID"), os.Getenv("TWITCH_CLIENT_SECRET"))
	if err != nil {
		panic(err)
	}
	apiWrapper.SetAppToken(appToken)
	apiWrapper.SetClientId(os.Getenv("TWITCH_CLIENT_ID"))

	return &PlayersManager{
		mutex:   &sync.Mutex{},
		clients: make(map[string]*websocket.Conn),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		eventSub:   twitch.GetEventSub(apiWrapper),
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

	notifcationHandler := twitch.GetNotificationHandler(pm.apiWrapper, token)

	pm.eventSub.OnEvent(token, func(event twitch.NotificationMessage) {
		notifcationHandler.Handle(&event)
	})

	if pm.eventSub.IsStarted() {
		pm.eventSub.SubscribeToMessageEvents(token)
	} else {
		pm.eventSub.OnStarted(func() {
			pm.eventSub.SubscribeToMessageEvents(token)
		})
		pm.eventSub.Start()
	}

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Client disconnected")
			conn.Close()

			pm.mutex.Lock()
			delete(pm.clients, token)
			pm.mutex.Unlock()

			pm.eventSub.DropAllSubscriptions(token)

			break
		}
	}
}
