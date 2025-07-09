package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Yaon-C2H8N2/bahclePlayer/internal/models"
	"github.com/Yaon-C2H8N2/bahclePlayer/internal/models/twitch"
	"github.com/Yaon-C2H8N2/bahclePlayer/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
	"os"
	"sync"
	"time"
)

var eventTypesHandler = map[string]func(*websocket.Conn, any){
	"currently_playing": handleEvent,
	"new_video":         handleEvent,
	"deleted_video":     handleEvent,
}

func getOverlays(c *gin.Context) {
	TwitchUserContext, _ := c.Get("TwitchUser")
	userInfo, _ := TwitchUserContext.(twitch.UserInfo)

	overlayTypes, err := models.GetAllOverlayTypes()
	if err != nil {
		c.JSON(500, gin.H{
			"error": "Failed to get overlay types",
		})
		return
	}

	var overlayTypesWithLinks []struct {
		models.OverlayType
		Link string `json:"link"`
	}
	for i, _ := range overlayTypes {
		var overlayTypeWithLink struct {
			models.OverlayType
			Link string `json:"link"`
		}
		overlayTypeWithLink.OverlayType = overlayTypes[i]
		overlayTypeWithLink.Link = os.Getenv("APP_URL") + "/overlay/" + userInfo.ID + "/" + overlayTypes[i].OverlayCode
		overlayTypesWithLinks = append(overlayTypesWithLinks, overlayTypeWithLink)
	}

	c.JSON(200, gin.H{
		"overlay_types": overlayTypesWithLinks,
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

func getUserOverlays(c *gin.Context) {
	userContext, _ := c.Get("User")
	user, _ := userContext.(models.Users)

	userOverlays, err := models.GetAllUsersOverlaysFromUserId(user.UserId)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "Failed to get user overlays",
		})
		return
	}

	c.JSON(200, gin.H{
		"user_overlays": userOverlays,
	})
}

func getUserOverlaySettings(c *gin.Context) {
	twitchId := c.Query("twitch_id")
	overlayCode := c.Query("overlay_code")

	if twitchId == "" || overlayCode == "" {
		c.JSON(400, gin.H{
			"error": "Missing required parameters: twitch_id and overlay_code",
		})
		return
	}

	userOverlaySettings, err := models.GetUserOverlaySettingsByTwitchId(twitchId, overlayCode)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "Failed to get user overlay settings",
		})
		return
	}

	c.JSON(200, gin.H{
		"settings": userOverlaySettings,
	})
}

func saveUserOverlaySettings(c *gin.Context) {
	userContext, _ := c.Get("User")
	user, _ := userContext.(models.Users)

	var request struct {
		OverlayCode string      `json:"overlay_code"`
		Settings    interface{} `json:"settings"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	err := models.SaveUserOverlaySettings(user.UserId, request.OverlayCode, request.Settings)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "Failed to save user overlay settings",
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "Settings saved successfully",
	})
}
