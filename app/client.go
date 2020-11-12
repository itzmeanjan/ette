package app

import (
	"log"

	"github.com/go-redis/redis/v8"

	"github.com/ethereum/go-ethereum/ethclient"
	cfg "github.com/itzmeanjan/ette/app/config"
)

// Connect to blockchain node
func getClient() *ethclient.Client {
	client, err := ethclient.Dial(cfg.Get("RPC"))
	if err != nil {
		log.Fatalf("[!] Failed to connect to blockchain : %s\n", err.Error())
	}

	return client
}

// Creates connection to redis server & returns that handle to be used for further communication
func getPubSubClient() *redis.Client {

	return redis.NewClient(&redis.Options{
		Network: cfg.Get("RedisConnection"),
		Addr:    cfg.Get("RedisAddress"),
		DB:      0,
	})

}
