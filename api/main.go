package main

import (
	"fmt"
	"github.com/Yaon-C2H8N2/bahclePlayer/endpoints"
	"os"
)
import "github.com/gin-gonic/gin"

func main() {
	router := gin.Default()

	endpoints.MapRoutes(router)
	apiPort := os.Getenv("API_PORT")

	err := router.Run(fmt.Sprintf(":%s", apiPort))
	if err != nil {
		panic(err)
	}
}
