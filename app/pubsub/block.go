package pubsub

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"

	"github.com/itzmeanjan/ette/app/data"
	"github.com/itzmeanjan/ette/app/db"
)

// BlockConsumer - To be subscribed to `block` topic using this consumer handle
// and client connected using websocket needs to be delivered this piece of data
type BlockConsumer struct {
	Client     *redis.Client
	Requests   map[string]*SubscriptionRequest
	Connection *websocket.Conn
	PubSub     *redis.PubSub
	DB         *gorm.DB
	ConnLock   *sync.Mutex
	TopicLock  *sync.RWMutex
	Counter    *data.SendReceiveCounter
}

// Subscribe - Subscribe to `block` channel
func (b *BlockConsumer) Subscribe() {
	b.PubSub = b.Client.Subscribe(context.Background(), "block")
}

// Listen - Listener function, which keeps looping in infinite loop
// and reads data from subcribed channel, which also gets delivered to client application
func (b *BlockConsumer) Listen() {

	for {

		msg, err := b.PubSub.ReceiveTimeout(context.Background(), time.Second)
		if err != nil {
			continue
		}

		switch m := msg.(type) {

		case *redis.Subscription:

			// Pubsub broker informed we've been unsubscribed from
			// this topic
			if m.Kind == "unsubscribe" {
				return
			}

			b.SendData(&SubscriptionResponse{
				Code:    1,
				Message: "Subscribed to `block`",
			})

		case *redis.Message:
			b.Send(m.Payload)

		}

	}

}

// Send - Tries to deliver subscribed block data to client application
// connected over websocket
func (b *BlockConsumer) Send(msg string) {

	var request *SubscriptionRequest

	// -- Shared memory being read from concurrently
	// running thread of execution, with lock
	b.TopicLock.RLock()

	for _, v := range b.Requests {

		request = v
		break

	}

	b.TopicLock.RUnlock()
	// -- Shared memory reading done, lock released

	// Can't proceed with this anymore, because failed to find
	// respective subscription request
	if request == nil {
		return
	}

	user := db.GetUserFromAPIKey(b.DB, request.APIKey)
	if user == nil {

		// -- Critical section of code begins
		//
		// Attempting to write to a network resource,
		// shared among multiple go routines
		b.ConnLock.Lock()

		if err := b.Connection.WriteJSON(&SubscriptionResponse{
			Code:    0,
			Message: "Bad API Key",
		}); err != nil {
			log.Printf("[!] Failed to deliver bad API key message to client : %s\n", err.Error())
		}

		b.ConnLock.Unlock()
		// -- ends here

		// Because we're writing to socket
		b.Counter.IncrementSend(1)
		return

	}

	if !user.Enabled {

		// -- Critical section of code begins
		//
		// Attempting to write to a network resource,
		// shared among multiple go routines
		b.ConnLock.Lock()

		if err := b.Connection.WriteJSON(&SubscriptionResponse{
			Code:    0,
			Message: "Bad API Key",
		}); err != nil {
			log.Printf("[!] Failed to deliver bad API key message to client : %s\n", err.Error())
		}

		b.ConnLock.Unlock()
		// -- ends here

		// Because we're writing to socket
		b.Counter.IncrementSend(1)
		return

	}

	// Don't deliver data & close underlying connection
	// if client has crossed it's allowed data delivery limit
	if !db.IsUnderRateLimit(b.DB, user.Address) {

		// -- Critical section of code begins
		//
		// Attempting to write to a network resource,
		// shared among multiple go routines
		b.ConnLock.Lock()

		if err := b.Connection.WriteJSON(&SubscriptionResponse{
			Code:    0,
			Message: "Crossed Allowed Rate Limit",
		}); err != nil {
			log.Printf("[!] Failed to deliver rate limit crossed message to client : %s\n", err.Error())
		}

		b.ConnLock.Unlock()
		// -- ends here

		// Because we're writing to socket
		b.Counter.IncrementSend(1)
		return

	}

	var block struct {
		Hash                string  `json:"hash"`
		Number              uint64  `json:"number"`
		Time                uint64  `json:"time"`
		ParentHash          string  `json:"parentHash"`
		Difficulty          string  `json:"difficulty"`
		GasUsed             uint64  `json:"gasUsed"`
		GasLimit            uint64  `json:"gasLimit"`
		Nonce               string  `json:"nonce"`
		Miner               string  `json:"miner"`
		Size                float64 `json:"size"`
		StateRootHash       string  `json:"stateRootHash"`
		UncleHash           string  `json:"uncleHash"`
		TransactionRootHash string  `json:"txRootHash"`
		ReceiptRootHash     string  `json:"receiptRootHash"`
		ExtraData           string  `json:"extraData"`
	}

	_msg := []byte(msg)

	err := json.Unmarshal(_msg, &block)
	if err != nil {
		log.Printf("[!] Failed to decode published block data to JSON : %s\n", err.Error())
		return
	}

	if b.SendData(&block) {
		db.PutDataDeliveryInfo(b.DB, user.Address, "/v1/ws/block", uint64(len(msg)))
	}

}

// SendData - Sending message to client application, connected over websocket
//
// If failed, we're going to remove subscription & close websocket
// connection ( connection might be already closed though )
func (b *BlockConsumer) SendData(data interface{}) bool {

	// -- Critical section of code begins
	//
	// Attempting to write to a network resource,
	// shared among multiple go routines
	b.ConnLock.Lock()
	defer b.ConnLock.Unlock()

	if err := b.Connection.WriteJSON(data); err != nil {
		log.Printf("[!] Failed to deliver `block` data to client : %s\n", err.Error())
		return false
	}

	// Because we're writing to socket
	b.Counter.IncrementSend(1)

	return true

}

// Unsubscribe - Unsubscribe from block data publishing event this client has subscribed to
func (b *BlockConsumer) Unsubscribe() {

	if b.PubSub == nil {
		log.Printf("[!] Bad attempt to unsubscribe from `block` topic\n")
		return
	}

	if err := b.PubSub.Unsubscribe(context.Background(), "block"); err != nil {
		log.Printf("[!] Failed to unsubscribe from `block` topic : %s\n", err.Error())
		return
	}

	resp := &SubscriptionResponse{
		Code:    1,
		Message: "Unsubscribed from `block`",
	}

	// -- Critical section of code begins
	//
	// Attempting to write to a network resource,
	// shared among multiple go routines
	b.ConnLock.Lock()
	defer b.ConnLock.Unlock()

	if err := b.Connection.WriteJSON(resp); err != nil {

		log.Printf("[!] Failed to deliver `block` unsubscription confirmation to client : %s\n", err.Error())
		return

	}

	// Because we're writing to socket
	b.Counter.IncrementSend(1)

}
