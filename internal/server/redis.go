package server

import (
	"context"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
)

var ctx = context.Background()
var redisClient *redis.Client

func InitRedis() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	redisPublicAddr := os.Getenv("REDIS_PUB_ADDR")
	redisPassword := os.Getenv("REDIS_PASSWORD")

	redisClient = redis.NewClient(&redis.Options{
		Addr:     redisPublicAddr,
		Password: redisPassword,
		DB:       0,
	})

	_, err = redisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
}
