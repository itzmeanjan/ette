package app

import (
	"context"
	"github.com/go-redis/redis/v8"
	cfg "github.com/itzmeanjan/ette/app/config"
)

// Creates connection to Redis server & returns that handle to be used for further communication
func getRedisClient() *redis.Client {

	var options *redis.Options

	// If password is given in config file
	if cfg.Get("RedisPassword") != "" {

		options = &redis.Options{
			Network:  cfg.Get("RedisConnection"),
			Addr:     cfg.Get("RedisAddress"),
			Password: cfg.Get("RedisPassword"),
			DB:       0,
		}

	} else {
		// If password is not given, attempting to connect with out it
		//
		// Though this is not recommended
		options = &redis.Options{
			Network: cfg.Get("RedisConnection"),
			Addr:    cfg.Get("RedisAddress"),
			DB:      0,
		}

	}

	_redis := redis.NewClient(options)
	// Checking whether connection was successful or not
	if err := _redis.Ping(context.Background()).Err(); err != nil {
		return nil
	}

	return _redis

}
