package utils

import (
	"fmt"

	"github.com/redis/go-redis/v9"
)

var client *redis.Client

func SetClient(redisClient *redis.Client){
	fmt.Println("Redis Client Set")
	client = redisClient
	fmt.Println("Redis Client is initialized")
}

func GetClient() *redis.Client{
	if client == nil{
		fmt.Println("Warning : Redis Client is nil")
	}
	return client
}
