package data

import (
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
)

// NewBlockConsumer - Creating one new block data consumer, which will subscribe to block
// topic & listen for data being published on this channel, which will eventually be
// delivered to client application over websocket connection
func NewBlockConsumer(client *redis.Client, conn *websocket.Conn, req *SubscriptionRequest) *BlockConsumer {
	consumer := BlockConsumer{
		Client:     client,
		Request:    req,
		Connection: conn,
	}

	consumer.Subscribe()
	go consumer.Listen()

	return &consumer
}

// NewTransactionConsumer - Creating one new transaction data consumer, which will subscribe to transaction
// topic & listen for data being published on this channel & check whether received data
// is what, client is interested in or not, which will eventually be
// delivered to client application over websocket connection
func NewTransactionConsumer(client *redis.Client, conn *websocket.Conn, req *SubscriptionRequest) *TransactionConsumer {
	consumer := TransactionConsumer{
		Client:     client,
		Request:    req,
		Connection: conn,
	}

	consumer.Subscribe()
	go consumer.Listen()

	return &consumer
}

// NewEventConsumer - Event data consumer, manages everything, starting from subscription to data reception
// delivery, also cleans up resources if connection gets closed
func NewEventConsumer(client *redis.Client, conn *websocket.Conn, req *SubscriptionRequest) *EventConsumer {
	consumer := EventConsumer{
		Client:     client,
		Request:    req,
		Connection: conn,
	}

	consumer.Subscribe()
	go consumer.Listen()

	return &consumer
}
