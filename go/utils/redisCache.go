package utils

import (
	"github.com/redis/go-redis/v9"
)

var rdb *redis.Client

func GetRedisClient() *redis.Client {
	return rdb
}

func ConnectRedisClient() {
	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
}
