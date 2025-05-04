package utils

import (
	"github.com/redis/go-redis/v9"
	"os"
)

var ValkeyClient *redis.Client

func InitValkey() {
	address := os.Getenv("VALKEY_URL")
	port := os.Getenv("VALKEY_PORT")

	ValkeyClient = redis.NewClient(&redis.Options{
		Addr:     address + ":" + port,
		Password: "",
		DB:       0,
	})
}

func GetValkeyClient() *redis.Client {
	return ValkeyClient
}
