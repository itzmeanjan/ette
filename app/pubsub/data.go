package pubsub

import (
	"fmt"
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"github.com/itzmeanjan/ette/app/data"
	"github.com/segmentio/kafka-go"
	"gorm.io/gorm"
)

// SubscriptionManager - Higher level abstraction to be used
// by websocket connection acceptor, for subscribing to topics
//
// They don't need to know that for each subscription request
// over same websocket connection, one new pubsub subscription
// may not be created
//
// For each client there could be possibly at max 3 pubsub subscriptions
// i.e. block, transaction, event, which are considered to be top level
// topics
//
// For each of them there could be multiple subtopics but not explicit
// pubsub subscription
//
// This is being done for reducing redundant pressure on pubsub
// broker i.e. Redis here 🥳
type SubscriptionManager struct {
	Topics     map[string]map[string]*SubscriptionRequest
	Consumers  map[string]Consumer
	Client     *redis.Client
	Connection *websocket.Conn
	DB         *gorm.DB
	ConnLock   *sync.Mutex
	TopicLock  *sync.RWMutex
	Counter    *data.SendReceiveCounter
}

// Subscribe - Websocket connection manager can reliably call
// this function when ever it receives one valid subscription request
// with out worrying about how will it be handled
func (s *SubscriptionManager) Subscribe(req *SubscriptionRequest, _kafkaWriter *kafka.Writer) {

	s.TopicLock.Lock()
	defer s.TopicLock.Unlock()

	_, ok := s.Topics[req.Topic()]
	if !ok {

		tmp := make(map[string]*SubscriptionRequest)
		tmp[req.Name] = req

		s.Topics[req.Topic()] = tmp

		switch req.Topic() {

		case "block":
			s.Consumers[req.Topic()] = NewBlockConsumer(s.Client, tmp, s.Connection, s.DB, s.ConnLock, s.TopicLock, s.Counter)
		case "transaction":
			s.Consumers[req.Topic()] = NewTransactionConsumer(s.Client, tmp, s.Connection, s.DB, s.ConnLock, s.TopicLock, s.Counter)
		case "event":
			s.Consumers[req.Topic()] = NewEventConsumer(s.Client, tmp, s.Connection, s.DB, s.ConnLock, s.TopicLock, s.Counter, _kafkaWriter)
		}

		return

	}

	s.Topics[req.Topic()][req.Name] = req
	s.Consumers[req.Topic()].SendData(
		&SubscriptionResponse{
			Code:    1,
			Message: fmt.Sprintf("Subscribed to `%s`", req.Topic()),
		})

}

// Unsubscribe - Websocket connection manager can reliably call
// this to unsubscribe from topic for this client
//
// If all subtopics for `block`/ `transaction`/ `event` are
// unsubscribed from, entry to be removed from associative array
// and pubsub to be unsubscribed
//
// Otherwise, we simply remove this specific topic from associative array
// holding subtopics for any of `block`/ `transaction`/ `event` root topics
func (s *SubscriptionManager) Unsubscribe(req *SubscriptionRequest) {

	s.TopicLock.Lock()
	defer s.TopicLock.Unlock()

	_, ok := s.Topics[req.Topic()]
	if !ok {
		return
	}

	delete(s.Topics[req.Topic()], req.Name)

	if len(s.Topics[req.Topic()]) > 0 {

		s.Consumers[req.Topic()].SendData(
			&SubscriptionResponse{
				Code:    1,
				Message: fmt.Sprintf("Unsubscribed from `%s`", req.Topic()),
			})
		return

	}

	s.Consumers[req.Topic()].Unsubscribe()
	delete(s.Topics, req.Topic())
	delete(s.Consumers, req.Topic())

}
