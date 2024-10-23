package endpoints

import (
	"github.com/gin-gonic/gin"
)

func login(c *gin.Context) {
	token := c.Query("access_token")

	if token == "" {
		c.JSON(400, gin.H{
			"error": "missing access_token",
		})
	}
	c.Header("Set-Cookie", "token="+token+"; Path=/;")
	c.JSON(200, gin.H{
		"token": token,
	})
}
