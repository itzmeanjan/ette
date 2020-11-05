package app

import (
	"log"

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

// Connect to redis & use connection for creating queues
func getRedisConnection() rmq.Connection {
	errChan := make(chan error)

	conn, err := rmq.OpenConnection("ette", cfg.Get("RedisConnection"), cfg.Get("RedisAddress"), 1, errChan)

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
