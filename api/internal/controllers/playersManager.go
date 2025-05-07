package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Yaon-C2H8N2/bahclePlayer/internal/models"
	"github.com/Yaon-C2H8N2/bahclePlayer/pkg/utils"
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
	apiWrapper *ApiWrapper
}

func DefaultPlayersManager(apiWrapper *ApiWrapper) *PlayersManager {
	return &PlayersManager{
		mutex:   &sync.Mutex{},
		clients: make(map[string]map[string]*websocket.Conn),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		apiWrapper: apiWrapper,
	}
}

func (pm *PlayersManager) GetConnFromTwitchId(twitchId string) map[string]*websocket.Conn {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	conn := pm.clients[twitchId]

	return conn
}

func sendErrorMessage(conn *websocket.Conn, message string) {
	payload := struct {
		Error   string `json:"error"`
		Message string `json:"message"`
	}{
		Error:   "Error",
		Message: message,
	}

	err := conn.WriteJSON(payload)
	if err != nil {
		conn.Close()
	}
}

func (pm *PlayersManager) closeAndDeleteConn(conn *websocket.Conn, twitchId string, connUuid string) {
	fmt.Println("Closing connection")
	conn.Close()
	pm.mutex.Lock()
	delete(pm.clients[twitchId], connUuid)
	pm.mutex.Unlock()
}

func (pm *PlayersManager) CreatePlayer(c *gin.Context) {
	conn, err := pm.upgrader.Upgrade(c.Writer, c.Request, nil)
	socketLock := &sync.Mutex{}

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
			socketLock.Lock()
			sendErrorMessage(conn, "No message received before timeout")
			conn.Close()
			socketLock.Unlock()
			break
		case <-tokenCh:
			break
		}
	}()

	stopWelcome := make(chan struct{})
	go func() {
		welcome := struct {
			Welcome string `json:"welcome"`
		}{
			Welcome: "Socket connection established",
		}

		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-stopWelcome:
				return
			case <-ticker.C:
				socketLock.Lock()
				err := conn.WriteJSON(welcome)
				socketLock.Unlock()
				if err != nil {
					return
				}
			}
		}
	}()

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

	jwtToken, err := models.ValidateToken(token)
	if err != nil {
		socketLock.Lock()
		sendErrorMessage(conn, "An error occured when validating token")
		conn.Close()
		socketLock.Unlock()
		return
	}

	tokenClaims := jwtToken.Claims.(*models.JWTClaims)
	user, err := models.GetUserFromUserId(tokenClaims.UserId)
	if err != nil {
		socketLock.Lock()
		sendErrorMessage(conn, "An error occured when getting user from token")
		conn.Close()
		socketLock.Unlock()
		return
	}

	userInfo, err := pm.apiWrapper.GetUserInfoFromToken(user.Token)
	if err != nil {
		socketLock.Lock()
		sendErrorMessage(conn, "An error occured when getting user info")
		conn.Close()
		socketLock.Unlock()
		return
	}

	pm.mutex.Lock()
	if token != "" {
		tokenCh <- token
		close(stopWelcome)

		if pm.clients[userInfo.ID] == nil {
			pm.clients[userInfo.ID] = make(map[string]*websocket.Conn, 0)
		}

		connUuid := uuid.New().String()
		pm.clients[userInfo.ID][connUuid] = conn

		go pm.mainLoop(userInfo.ID, connUuid)
	}
	pm.mutex.Unlock()
}

func (pm *PlayersManager) onNewVideo(conn *websocket.Conn, newVideo models.UsersVideos) {
	err := conn.WriteJSON(newVideo)
	if err != nil {
		sendErrorMessage(conn, "An error occured when sending new video")
	}
}

func (pm *PlayersManager) mainLoop(twitchId string, connUuid string) {
	conn := pm.clients[twitchId][connUuid]
	valkeyClient := utils.GetValkeyClient()
	sub := valkeyClient.Subscribe(context.Background(), "eventsub:"+twitchId+":new_video")
	defer sub.Close()

	// Set the pong handler to reset the read deadline when a pong message is received
	conn.SetPongHandler(func(appData string) error {
		return conn.SetReadDeadline(time.Now().Add(15 * time.Second))
	})
	_ = conn.SetReadDeadline(time.Now().Add(15 * time.Second))

	// Keep reading messages from the connection to listen for pong responses
	go func() {
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				pm.closeAndDeleteConn(conn, twitchId, connUuid)
				return
			}
		}
	}()

	go func() {
		for msg := range sub.Channel() {
			var newVideo models.UsersVideos
			err := json.Unmarshal([]byte(msg.Payload), &newVideo)
			if err != nil {
				sendErrorMessage(conn, "An error occured when unmarshalling new video")
				continue
			}
			pm.onNewVideo(conn, newVideo)
		}
	}()

	// Send ping messages every 5 seconds to keep the connection alive
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
			if err := conn.WriteMessage(websocket.PingMessage, []byte("ping")); err != nil {
				pm.closeAndDeleteConn(conn, twitchId, connUuid)
				return
			}
		}
	}
}
