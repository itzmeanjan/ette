package pubsub

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

// Consumer - Block, transaction & event consumers need to implement these methods
type Consumer interface {
	Subscribe()
	Listen()
	Send(msg string) bool
	SendData(data interface{}) bool
}

// NewBlockConsumer - Creating one new block data consumer, which will subscribe to block
// topic & listen for data being published on this channel, which will eventually be
// delivered to client application over websocket connection
func NewBlockConsumer(client *redis.Client, conn *websocket.Conn, req *SubscriptionRequest, db *gorm.DB, address common.Address) *BlockConsumer {
	consumer := BlockConsumer{
		Client:      client,
		Request:     req,
		UserAddress: address,
		Connection:  conn,
		DB:          db,
	}

	consumer.Subscribe()
	go consumer.Listen()

	return &consumer
}

// NewTransactionConsumer - Creating one new transaction data consumer, which will subscribe to transaction
// topic & listen for data being published on this channel & check whether received data
// is what, client is interested in or not, which will eventually be
// delivered to client application over websocket connection
func NewTransactionConsumer(client *redis.Client, conn *websocket.Conn, req *SubscriptionRequest, db *gorm.DB, address common.Address) *TransactionConsumer {
	consumer := TransactionConsumer{
		Client:      client,
		Request:     req,
		UserAddress: address,
		Connection:  conn,
		DB:          db,
	}

	consumer.Subscribe()
	go consumer.Listen()

	return &consumer
}

// NewEventConsumer - Creating one new event data consumer, which will subscribe to event
// topic & listen for data being published on this channel & check whether received data
// is what, client is interested in or not, which will eventually be
// delivered to client application over websocket connection
func NewEventConsumer(client *redis.Client, conn *websocket.Conn, req *SubscriptionRequest, db *gorm.DB, address common.Address) *EventConsumer {
	consumer := EventConsumer{
		Client:      client,
		Request:     req,
		UserAddress: address,
		Connection:  conn,
		DB:          db,
	}

	consumer.Subscribe()
	go consumer.Listen()

	return &consumer
}
