package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	internalContext "github.com/Yaon-C2H8N2/bahclePlayer/internal/context"
	"github.com/Yaon-C2H8N2/bahclePlayer/internal/controllers"
	"github.com/Yaon-C2H8N2/bahclePlayer/internal/models"
	"github.com/Yaon-C2H8N2/bahclePlayer/internal/router"
	"github.com/Yaon-C2H8N2/bahclePlayer/internal/services"
	"github.com/Yaon-C2H8N2/bahclePlayer/pkg/utils"
)
import "github.com/gin-gonic/gin"

func main() {
	utils.InitDatabase()
	utils.InitValkey()

	valkeyClient := utils.GetValkeyClient()
	sub := valkeyClient.PSubscribe(context.Background(), "__keyevent@0__:expired")
	fmt.Println("Subscribed to keyevent:expired")

	appStatus := models.AppStatus{
		TwitchClientId: os.Getenv("TWITCH_CLIENT_ID"),
		AppUrl:         os.Getenv("APP_URL"),
		Started:        false,
		Version:        os.Getenv("APP_VERSION"),
	}

	appToken, appTokenErr := controllers.RequestAppToken(appStatus.AppUrl, os.Getenv("TWITCH_CLIENT_SECRET"))
	if appTokenErr != nil {
		panic(appTokenErr)
	}
	apiWrapper := controllers.GetApiWrapper(appToken, os.Getenv("TWITCH_CLIENT_ID"))

	users, err := models.GetAllUsers()
	if err != nil {
		panic(err)
	}

	eventSubPool := controllers.GetEventSubPool(os.Getenv("TWITCH_EVENTSUB_WEBSOCKET_URL"), apiWrapper)
	for _, user := range users {
		if user.Token == "" {
			continue
		}
		fmt.Printf("Initializing subscriptions for user %s\n", user.Username)
		err = eventSubPool.AddEventSub(user)
		if err != nil {
			fmt.Printf("Error initializing EventSub for user %s: %s\n", user.Username, err)
			continue
		}
	}

	fmt.Println("EventSubs initialized")
	playersManager := controllers.DefaultPlayersManager(apiWrapper)

	ginRouter := gin.New()
	ginRouter.Use(func(c *gin.Context) {
		router.AuthMiddleware(c, apiWrapper)
	}, gin.Recovery())

	appContext := &internalContext.AppContext{
		ApiWrapper:     apiWrapper,
		EventSubPool:   eventSubPool,
		PlayersManager: playersManager,
		AppStatus:      &appStatus,
	}
	servicesList := []any{
		services.GetLoginService(),
		services.GetPlayerService(),
		services.GetOverlaysService(),
		services.GetSettingsService(),
	}
	for _, service := range servicesList {
		router.RegisterService(ginRouter, appContext, service)
	}
	ginRouter.GET("/appinfo", func(c *gin.Context) {
		c.JSON(200, appContext.AppStatus)
	})

	go func() {
		for msg := range sub.Channel() {
			if !strings.HasPrefix(msg.Payload, "auth:token:") {
				continue
			}

			oldToken := msg.Payload[len("auth:token:"):]

			fmt.Println("Token expired:", oldToken)
			user, err := models.GetUserFromToken(oldToken)
			if err != nil {
				fmt.Println("Error getting token:", err)
				continue
			}
			refreshedToken, err := controllers.RefreshUserToken(user.RefreshToken)
			if err != nil {
				fmt.Println("Error refreshing token:", err)
				continue
			}

			newUser, err := models.AddOrUpdateUser(user, *refreshedToken)
			if err != nil {
				fmt.Println("Error updating user:", err)
				continue
			}

			err = eventSubPool.UpdateUser(newUser)
			if err != nil {
				fmt.Println("Error updating EventSub for user:", err)
				continue
			}
			fmt.Println("Token refreshed for user:", newUser.TwitchId)
		}
		fmt.Println("Keyevent listener stopped")
	}()

	appStatus.Started = true
	err = ginRouter.Run(fmt.Sprintf(":%d", 8081))
	if err != nil {
		panic(err)
	}
}
