package models

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"os"
	"time"
)

type JWTClaims struct {
	UserId   int    `json:"user_id"`
	TwitchId string `json:"twitch_id"`
	Username string `json:"username"`
	Exp      int64  `json:"exp"`
	jwt.RegisteredClaims
}

func GenerateToken(user Users) (string, error) {
	secretKey := os.Getenv("JWT_SECRET_KEY")
	secretKeyBytes := []byte(secretKey)

	claims := JWTClaims{
		UserId:   user.UserId,
		TwitchId: user.TwitchId,
		Username: user.Username,
		Exp:      time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(secretKeyBytes)
}

func ValidateToken(tokenString string) (*jwt.Token, error) {
	secretKey := os.Getenv("JWT_SECRET_KEY")

	return jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})
}
