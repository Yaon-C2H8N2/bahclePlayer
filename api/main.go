package main

import "fmt"
import "github.com/gin-gonic/gin"

func main() {
	router := gin.Default()

	err := router.Run(":8081")
	if err != nil {
		panic(err)
	}
	fmt.Println("Hello, World!")
}
