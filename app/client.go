package app

import (
	"log"

	"github.com/go-redis/redis/v8"

	rmq "github.com/adjust/rmq/v3"
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

// Connect to redis & use connection for creating queues
func getRedisConnection() rmq.Connection {
	errChan := make(chan error)

	conn, err := rmq.OpenConnection("ette", cfg.Get("RedisConnectionType"), cfg.Get("RedisAddress"), 1, errChan)

	if err != nil {
		log.Fatalf("[!] Failed to connect to redis : %s\n", err.Error())
	}

	return conn
}

// Create a redis message queue, using given db connection
func getRedisMessageQueue(connection rmq.Connection, queue string) rmq.Queue {
	_queue, err := connection.OpenQueue(queue)
	if err != nil {
		log.Fatalf("[!] Failed to create queue : `%s`\n", queue)
	}

	return _queue
}
