package main

import (
	"fmt"
	"github.com/Yaon-C2H8N2/bahclePlayer/internal/controllers"
	"github.com/Yaon-C2H8N2/bahclePlayer/internal/models"
	"github.com/Yaon-C2H8N2/bahclePlayer/internal/services"
	"github.com/Yaon-C2H8N2/bahclePlayer/pkg/utils"
	"os"
)
import "github.com/gin-gonic/gin"

func main() {
	utils.Migrate()
	utils.InitDatabase()

	router := gin.Default()

	appStatus := models.AppStatus{
		TwitchClientId: os.Getenv("TWITCH_CLIENT_ID"),
		AppUrl:         os.Getenv("APP_URL"),
		Started:        false,
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

	services.MapRoutes(router, playersManager, apiWrapper, eventSubs, &appStatus)
	apiPort := os.Getenv("API_PORT")

	appStatus.Started = true
	err := router.Run(fmt.Sprintf(":%s", apiPort))
	if err != nil {
		panic(err)
	}
}
