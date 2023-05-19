package infra

import (
	"github.com/redis/go-redis/v9"
	"voice-dispatch/config"
)

var RedisClient *redis.Client

func RedisInit() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     config.AppConfig.Redis.Addr,
		Password: config.AppConfig.Redis.Password, // no password set
		DB:       config.AppConfig.Redis.DB,       // use default DB
	})
}
