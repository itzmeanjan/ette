package pubsub

import (
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
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
	Topics      map[string]map[string]bool
	Consumers   map[string]Consumer
	Client      *redis.Client
	UserAddress common.Address
	Connection  *websocket.Conn
	PubSub      *redis.PubSub
	DB          *gorm.DB
	ConnLock    *sync.Mutex
	TopicLock   *sync.RWMutex
}

// Subscribe - Websocket connection manager can reliably call
// this function when ever it receives one valid subscription request
// with out worrying about how will it be handled
func (s *SubscriptionManager) Subscribe(req *SubscriptionRequest) {

	s.TopicLock.Lock()
	defer s.TopicLock.Unlock()

	_, ok := s.Topics[req.Topic()]
	if !ok {

		tmp := make(map[string]bool)
		tmp[req.Name] = true

		s.Topics[req.Topic()] = tmp

		switch req.Topic() {

		case "block":

			s.Consumers[req.Topic()] = NewBlockConsumer(s.Client, s.Connection, req, s.DB, s.UserAddress, s.ConnLock)

		case "transaction":

			s.Consumers[req.Topic()] = NewTransactionConsumer(s.Client, s.Connection, req, s.DB, s.UserAddress, s.ConnLock)

		case "event":

			s.Consumers[req.Topic()] = NewEventConsumer(s.Client, s.Connection, req, s.DB, s.UserAddress, s.ConnLock)

		}

		return

	}

	s.Topics[req.Topic()][req.Name] = true
	s.Consumers[req.Topic()].SendData(
		&SubscriptionResponse{
			Code:    1,
			Message: fmt.Sprintf("Subscribed to `%s`", req.Topic()),
		})

}
