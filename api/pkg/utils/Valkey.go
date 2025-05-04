package utils

import (
	"context"
	"github.com/redis/go-redis/v9"
	"os"
)

var ValkeyClient *redis.Client

func InitValkey() {
	address := os.Getenv("VALKEY_URL")
	port := os.Getenv("VALKEY_PORT")
	password := os.Getenv("VALKEY_PASSWORD")

	ValkeyClient = redis.NewClient(&redis.Options{
		Addr:     address + ":" + port,
		Password: password,
		DB:       0,
	})

	_, err := ValkeyClient.Ping(context.Background()).Result()
	if err != nil {
		panic(err)
	}
}

func GetValkeyClient() *redis.Client {
	return ValkeyClient
}
