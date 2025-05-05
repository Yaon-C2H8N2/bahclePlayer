package main

import (
	"context"
	"fmt"
	"github.com/Yaon-C2H8N2/bahclePlayer/internal/controllers"
	"github.com/Yaon-C2H8N2/bahclePlayer/internal/models"
	"github.com/Yaon-C2H8N2/bahclePlayer/internal/services"
	"github.com/Yaon-C2H8N2/bahclePlayer/pkg/utils"
	"os"
	"strings"
)
import "github.com/gin-gonic/gin"

func main() {
	if os.Getenv("API_ENV") == "" {
		err := utils.LoadEnv("../.env.development")
		if err != nil {
			panic("Failed to load env file")
		}
	}

	utils.Migrate()
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
	apiWrapper := controllers.GetApiWrapper()
	apiWrapper.SetClientId(os.Getenv("TWITCH_CLIENT_ID"))
	apiWrapper.SetAppToken(appToken)

	eventSubs := controllers.GetForAllUsers(apiWrapper)
	fmt.Println("EventSubs initialized")
	playersManager := controllers.DefaultPlayersManager(eventSubs, apiWrapper)

	router := gin.Default()
	router.Use(func(c *gin.Context) {
		services.AuthMiddleware(c, apiWrapper)
	})
	services.MapRoutes(router, playersManager, apiWrapper, eventSubs, &appStatus)

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

			eventSub := eventSubs[user.TwitchId]
			if eventSub == nil {
				fmt.Println("Event sub not found for user:", user.TwitchId)
				continue
			}
			eventSubs[user.TwitchId].UpdateUser(newUser)
		}
		fmt.Println("Keyevent listener stopped")
	}()

	appStatus.Started = true
	err := router.Run(fmt.Sprintf(":%d", 8081))
	if err != nil {
		panic(err)
	}
}
