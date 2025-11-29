package services

import (
	"github.com/Yaon-C2H8N2/bahclePlayer/internal/context"
	"github.com/Yaon-C2H8N2/bahclePlayer/internal/controllers"
	"github.com/Yaon-C2H8N2/bahclePlayer/internal/models"
	"github.com/Yaon-C2H8N2/bahclePlayer/internal/router"
	"github.com/Yaon-C2H8N2/bahclePlayer/pkg/utils"
	"github.com/gin-gonic/gin"
)

type LoginService struct {
	Login  router.HandlerFunction `method:"POST" path:"/login"`
	Logout router.HandlerFunction `method:"GET" path:"/logout"`
}

func GetLoginService() *LoginService {
	return &LoginService{
		Login:  login,
		Logout: logout,
	}
}

func login(c *gin.Context, appContext *context.AppContext) {
	loginRequest := models.LoginRequest{}
	err := c.BindJSON(&loginRequest)

	if err != nil {
		c.JSON(400, gin.H{
			"error": "Failed to bind login request",
		})
		return
	}

	conn := utils.GetConnection()
	sql := `
			INSERT INTO token_requests (code, requested_at)
			VALUES ($1, now())
			ON CONFLICT (code) DO NOTHING
			RETURNING code
		`
	rows := utils.DoRequest(conn, sql, loginRequest.Code)
	var user models.Users
	err = rows.Err()
	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}
	if !rows.Next() {
		c.JSON(401, gin.H{
			"error": "A token request with this code already exists",
		})
		return
	} else {
		userToken, err := controllers.RequestUserToken(loginRequest.Code)
		if err != nil {
			c.JSON(500, gin.H{
				"error": err.Error(),
			})
			return
		}

		conn.Release()
		conn := utils.GetConnection()
		sql = `
			DELETE FROM token_requests
			WHERE code = $1
		`
		utils.DoRequest(conn, sql, loginRequest.Code)

		userInfo, err := appContext.ApiWrapper.GetUserInfoFromToken(userToken.AccessToken)
		if err != nil {
			c.JSON(500, gin.H{
				"error": err.Error(),
			})
			return
		}
		conn.Release()

		user, err = models.AddOrUpdateUser(models.Users{
			Username:     userInfo.DisplayName,
			TwitchId:     userInfo.ID,
			Token:        userToken.AccessToken,
			RefreshToken: userToken.RefreshToken,
		}, *userToken)

		if err != nil {
			c.JSON(500, gin.H{
				"error": err.Error(),
			})
			return
		}
		err = appContext.EventSubPool.AddEventSub(appContext.ApiWrapper, user)
		if err != nil {
			c.JSON(500, gin.H{
				"error": "Failed to initialize EventSub: " + err.Error(),
			})
			return
		}
	}

	jwtToken, err := models.GenerateToken(user)
	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.Header("Set-Cookie", "token="+jwtToken+"; Path=/;")
	c.JSON(200, gin.H{
		"token": jwtToken,
		"user":  user,
	})
}

func logout(c *gin.Context, appContext *context.AppContext) {
	c.Header("Set-Cookie", "token=; Path=/; Expires=Thu, 01 Jan 1970 00:00:00 GMT")
	c.JSON(200, gin.H{
		"message": "Logged out",
	})
}
