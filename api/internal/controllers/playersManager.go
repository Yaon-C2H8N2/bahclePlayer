package controllers

import (
	"encoding/json"
	"github.com/Yaon-C2H8N2/bahclePlayer/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
	"time"
)

type PlayersManager struct {
	mutex      *sync.Mutex
	clients    map[string]map[string]*websocket.Conn
	upgrader   websocket.Upgrader
	eventSubs  map[string]*EventSub
	apiWrapper *ApiWrapper
}

func DefaultPlayersManager(eventSubs map[string]*EventSub, apiWrapper *ApiWrapper) *PlayersManager {
	return &PlayersManager{
		mutex:   &sync.Mutex{},
		clients: make(map[string]map[string]*websocket.Conn),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		eventSubs:  eventSubs,
		apiWrapper: apiWrapper,
	}
}

func (pm *PlayersManager) GetConnFromTwitchId(twitchId string) map[string]*websocket.Conn {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	conn := pm.clients[twitchId]

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

	tokenCh := make(chan string, 1)
	go func() {
		select {
		case <-time.After(10 * time.Second):
			payload := struct {
				Error   string `json:"error"`
				Message string `json:"message"`
			}{
				Error:   "Timeout",
				Message: "No message received before timeout",
			}

			conn.WriteJSON(payload)
			conn.Close()
			break
		case <-tokenCh:
			break
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
		conn.Close()
		return
	}

	var message struct {
		Token string `json:"token"`
	}
	json.Unmarshal(messageBytes, &message)
	token := message.Token

	userInfo, err := pm.apiWrapper.GetUserInfoFromToken(token)
	if err != nil {
		payload := struct {
			Error   string `json:"error"`
			Message string `json:"message"`
		}{
			Error:   err.Error(),
			Message: "An error occured when getting user info",
		}
		conn.WriteJSON(payload)
		conn.Close()
		return
	}

	if token != "" {
		tokenCh <- token

		if pm.clients[userInfo.ID] == nil {
			pm.clients[userInfo.ID] = make(map[string]*websocket.Conn, 0)
		}

		connUuid := uuid.New().String()
		pm.clients[userInfo.ID][connUuid] = conn

		go pm.mainLoop(userInfo.ID, connUuid)
	}
	pm.mutex.Unlock()
}

func (pm *PlayersManager) mainLoop(twitchId string, connUuid string) {
	conn := pm.clients[twitchId][connUuid]
	eventSub := pm.eventSubs[twitchId]

	unsubscribeEvent := eventSub.notificationHandler.OnNewVideo(func(newVideo models.UsersVideos) {
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
	//unsubscribeMessage := eventSub.notificationHandler.OnChatMessage(func(message twitch.ChatMessageEvent) {
	//	err := conn.WriteJSON(message)
	//	if err != nil {
	//		payload := struct {
	//			Error   string `json:"error"`
	//			Message string `json:"message"`
	//		}{
	//			Error:   err.Error(),
	//			Message: "An error occured when sending new video",
	//		}
	//
	//		conn.WriteJSON(payload)
	//	}
	//})

	for {
		err := conn.WriteControl(websocket.PingMessage, []byte("PING!"), time.Now().Add(5*time.Second))
		if err != nil {
			unsubscribeEvent()
			unsubscribeError()
			//unsubscribeMessage()
			conn.Close()

			pm.mutex.Lock()
			delete(pm.clients[twitchId], connUuid)
			pm.mutex.Unlock()

			break
		}
		time.Sleep(5 * time.Second)
	}
}
