package pubsub

import (
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"github.com/itzmeanjan/ette/app/data"
	"github.com/segmentio/kafka-go"
	"gorm.io/gorm"
)

// Consumer - Block, transaction & event consumers need to implement these methods
type Consumer interface {
	Subscribe()
	Listen()
	Send(msg string)
	SendData(data interface{}) bool
	Unsubscribe()
}

// NewBlockConsumer - Creating one new block data consumer, which will subscribe to block
// topic & listen for data being published on this channel, which will eventually be
// delivered to client application over websocket connection
func NewBlockConsumer(client *redis.Client, requests map[string]*SubscriptionRequest, conn *websocket.Conn, db *gorm.DB, connLock *sync.Mutex, topicLock *sync.RWMutex, counter *data.SendReceiveCounter) *BlockConsumer {
	consumer := BlockConsumer{
		Client:     client,
		Requests:   requests,
		Connection: conn,
		DB:         db,
		ConnLock:   connLock,
		TopicLock:  topicLock,
		Counter:    counter,
	}

	consumer.Subscribe()
	go consumer.Listen()

	return &consumer
}

// NewTransactionConsumer - Creating one new transaction data consumer, which will subscribe to transaction
// topic & listen for data being published on this channel & check whether received data
// is what, client is interested in or not, which will eventually be
// delivered to client application over websocket connection
func NewTransactionConsumer(client *redis.Client, requests map[string]*SubscriptionRequest, conn *websocket.Conn, db *gorm.DB, connLock *sync.Mutex, topicLock *sync.RWMutex, counter *data.SendReceiveCounter) *TransactionConsumer {
	consumer := TransactionConsumer{
		Client:     client,
		Requests:   requests,
		Connection: conn,
		DB:         db,
		ConnLock:   connLock,
		TopicLock:  topicLock,
		Counter:    counter,
	}

	consumer.Subscribe()
	go consumer.Listen()

	return &consumer
}

// NewEventConsumer - Creating one new event data consumer, which will subscribe to event
// topic & listen for data being published on this channel & check whether received data
// is what, client is interested in or not, which will eventually be
// delivered to client application over websocket connection
func NewEventConsumer(client *redis.Client, requests map[string]*SubscriptionRequest, conn *websocket.Conn, db *gorm.DB, connLock *sync.Mutex, topicLock *sync.RWMutex, counter *data.SendReceiveCounter, _kafkaWriter *kafka.Writer) *EventConsumer {
	consumer := EventConsumer{
		Client:      client,
		Requests:    requests,
		Connection:  conn,
		DB:          db,
		ConnLock:    connLock,
		TopicLock:   topicLock,
		Counter:     counter,
		KafkaWriter: _kafkaWriter,
	}

	consumer.Subscribe()
	go consumer.Listen()

	return &consumer
}
