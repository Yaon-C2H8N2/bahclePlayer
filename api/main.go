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

	eventSubs := controllers.GetForAllUsers(apiWrapper)
	fmt.Println("EventSubs initialized")
	playersManager := controllers.DefaultPlayersManager(eventSubs, apiWrapper)

	endpoints.MapRoutes(router, playersManager, apiWrapper, eventSubs)
	apiPort := os.Getenv("API_PORT")

	err := router.Run(fmt.Sprintf(":%s", apiPort))
	if err != nil {
		panic(err)
	}
}
