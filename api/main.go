package main

import (
	"fmt"
	"github.com/Yaon-C2H8N2/bahclePlayer/endpoints"
	"github.com/Yaon-C2H8N2/bahclePlayer/utils"
	"os"
)
import "github.com/gin-gonic/gin"

func main() {
	utils.Migrate()

	router := gin.Default()

	endpoints.MapRoutes(router)
	apiPort := os.Getenv("API_PORT")

	err := router.Run(fmt.Sprintf(":%s", apiPort))
	if err != nil {
		panic(err)
	}
}
