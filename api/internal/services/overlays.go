package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Yaon-C2H8N2/bahclePlayer/internal/models"
	"github.com/Yaon-C2H8N2/bahclePlayer/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
	"time"
)

var eventTypesHandler = map[string]func(*websocket.Conn, any){
	"currently_playing": handleEvent,
	"new_video":         handleEvent,
	"deleted_video":     handleEvent,
}

func getOverlays(c *gin.Context) {
	conn := utils.GetConnection()
	defer conn.Release()

	sql := `
		SELECT overlay_type_id, name, description, schema
		FROM overlay_types
	`
	rows := utils.DoRequest(conn, sql)
	var overlayTypes []models.OverlayType
	for rows.Next() {
		var overlayType models.OverlayType
		err := rows.Scan(&overlayType.OverlayTypeId, &overlayType.Name, &overlayType.Description, &overlayType.Schema)
		if err != nil {
			fmt.Println("Failed to get overlay types:", err)
			c.JSON(500, gin.H{
				"error": "Failed to get overlay types",
			})
			return
		}
		overlayTypes = append(overlayTypes, overlayType)
	}

	c.JSON(200, gin.H{
		"overlay_types": overlayTypes,
	})
}

func handleEvent(conn *websocket.Conn, data any) {
	err := conn.WriteJSON(data)
	if err != nil {
		fmt.Println("Failed to send event data:", err)
	}
}

func getEventSocket(c *gin.Context) {
	twitchID := c.Query("twitch_id")
	eventType := c.Query("event_type")
	if twitchID == "" || eventType == "" {
		c.JSON(400, gin.H{
			"message": "Missing required parameters: twitch_id and event_type",
		})
		return
	}
	if _, ok := eventTypesHandler[eventType]; !ok {
		c.JSON(400, gin.H{
			"message": "Invalid event type",
		})
		return
	}

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(500, gin.H{
			"message": "Failed to upgrade connection",
			"error":   err.Error(),
		})
		return
	}
	socketLock := &sync.Mutex{}

	defer conn.Close()
	valkeyClient := utils.GetValkeyClient()
	sub := valkeyClient.Subscribe(context.Background(), "player:"+twitchID+":"+eventType)
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
				return
			}
		}
	}()

	go func() {
		for msg := range sub.Channel() {
			var data any
			err := json.Unmarshal([]byte(msg.Payload), &data)
			if err != nil {
				fmt.Println("Failed to unmarshal event data:", err)
				continue
			}
			socketLock.Lock()
			eventTypesHandler[eventType](conn, data)
			socketLock.Unlock()
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
				return
			}
		}
	}
}
