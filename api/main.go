package main

import (
	"fmt"
	"github.com/Yaon-C2H8N2/bahclePlayer/controllers"
	"github.com/Yaon-C2H8N2/bahclePlayer/endpoints"
	"github.com/Yaon-C2H8N2/bahclePlayer/utils"
	"os"
)
import "github.com/gin-gonic/gin"

func main() {
	utils.Migrate()

	router := gin.Default()

	appToken, appTokenErr := controllers.RequestAppToken(os.Getenv("TWITCH_CLIENT_ID"), os.Getenv("TWITCH_CLIENT_SECRET"))
	if appTokenErr != nil {
		panic(appTokenErr)
	}
	apiWrapper := controllers.GetApiWrapper()
	apiWrapper.SetClientId(os.Getenv("TWITCH_CLIENT_ID"))
	apiWrapper.SetAppToken(appToken)

	eventSub := controllers.GetEventSub(apiWrapper)
	eventSub.OnStarted(func() {
		fmt.Println("EventSub listener started")
		eventSub.InitForAllUsers()
		fmt.Println("EventSub subscriptions initialized")
	})
	eventSub.Start()

	playersManager := controllers.DefaultPlayersManager(eventSub)

	endpoints.MapRoutes(router, playersManager, apiWrapper, eventSub)
	apiPort := os.Getenv("API_PORT")

	err := router.Run(fmt.Sprintf(":%s", apiPort))
	if err != nil {
		panic(err)
	}
}
