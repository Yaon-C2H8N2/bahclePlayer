package main

import (
	"github.com/Yaon-C2H8N2/bahclePlayer/endpoints"
)
import "github.com/gin-gonic/gin"

func main() {
	router := gin.Default()

	endpoints.MapRoutes(router)

	err := router.Run(":8081")
	if err != nil {
		panic(err)
	}
}
