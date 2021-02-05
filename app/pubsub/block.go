package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"

	"github.com/itzmeanjan/ette/app/db"
)

// BlockConsumer - To be subscribed to `block` topic using this consumer handle
// and client connected using websocket needs to be delivered this piece of data
type BlockConsumer struct {
	Client     *redis.Client
	Request    *SubscriptionRequest
	Connection *websocket.Conn
	PubSub     *redis.PubSub
	DB         *gorm.DB
	Lock       *sync.Mutex
}

// Subscribe - Subscribe to `block` channel
func (b *BlockConsumer) Subscribe() {
	b.PubSub = b.Client.Subscribe(context.Background(), b.Request.Topic())
}

// Listen - Listener function, which keeps looping in infinite loop
// and reads data from subcribed channel, which also gets delivered to client application
func (b *BlockConsumer) Listen() {

	// When ever returning from this function's
	// execution context, client will be unsubscribed from
	// pubsub topic i.e. `block` topic in this case
	defer b.Unsubscribe()

	for {

		// Checking if client is still subscribed to this topic
		// or not
		//
		// If not, we're cancelling this subscription
		if b.Request.Type == "unsubscribe" {
			break
		}

		msg, err := b.PubSub.ReceiveTimeout(context.Background(), time.Second)
		if err != nil {
			continue
		}

		// To be used for checking whether delivering data to client went successful or not
		status := true

		switch m := msg.(type) {

		case *redis.Subscription:

			if m.Kind == "unsubscribe" {
				status = false
				break
			}

			status = b.SendData(&SubscriptionResponse{
				Code:    1,
				Message: "Subscribed to `block`",
			})

		case *redis.Message:
			status = b.Send(m.Payload)

		}

		if !status {
			break
		}
	}

}

// Send - Tries to deliver subscribed block data to client application
// connected over websocket
func (b *BlockConsumer) Send(msg string) bool {

	user := db.GetUserFromAPIKey(b.DB, b.Request.APIKey)
	if user == nil {

		// -- Critical section of code begins
		//
		// Attempting to write to a network resource,
		// shared among multiple go routines
		b.Lock.Lock()

		if err := b.Connection.WriteJSON(&SubscriptionResponse{
			Code:    0,
			Message: "Bad API Key",
		}); err != nil {
			log.Printf("[!] Failed to deliver bad API key message to client : %s\n", err.Error())
		}

		b.Lock.Unlock()
		// -- ends here
		return false

	}

	if !user.Enabled {

		// -- Critical section of code begins
		//
		// Attempting to write to a network resource,
		// shared among multiple go routines
		b.Lock.Lock()

		if err := b.Connection.WriteJSON(&SubscriptionResponse{
			Code:    0,
			Message: "Bad API Key",
		}); err != nil {
			log.Printf("[!] Failed to deliver bad API key message to client : %s\n", err.Error())
		}

		b.Lock.Unlock()
		// -- ends here
		return false

	}

	// Don't deliver data & close underlying connection
	// if client has crossed it's allowed data delivery limit
	if !db.IsUnderRateLimit(b.DB, user.Address) {

		// -- Critical section of code begins
		//
		// Attempting to write to a network resource,
		// shared among multiple go routines
		b.Lock.Lock()

		if err := b.Connection.WriteJSON(&SubscriptionResponse{
			Code:    0,
			Message: "Crossed Allowed Rate Limit",
		}); err != nil {
			log.Printf("[!] Failed to deliver rate limit crossed message to client : %s\n", err.Error())
		}

		b.Lock.Unlock()
		// -- ends here
		return false

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
		return true
	}

	if b.SendData(&block) {
		db.PutDataDeliveryInfo(b.DB, user.Address, "/v1/ws/block", uint64(len(msg)))
		return true
	}

	return false

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
	b.Lock.Lock()
	defer b.Lock.Unlock()

	if err := b.Connection.WriteJSON(data); err != nil {
		log.Printf("[!] Failed to deliver `block` data to client : %s\n", err.Error())
		return false
	}

	return true

}

// Unsubscribe - Unsubscribe from block data publishing event this client has subscribed to
func (b *BlockConsumer) Unsubscribe() {

	if b.PubSub == nil {
		log.Printf("[!] Bad attempt to unsubscribe from `%s` topic\n", b.Request.Topic())
		return
	}

	if err := b.PubSub.Unsubscribe(context.Background(), b.Request.Topic()); err != nil {
		log.Printf("[!] Failed to unsubscribe from `%s` topic : %s\n", b.Request.Topic(), err.Error())
		return
	}

	resp := &SubscriptionResponse{
		Code:    1,
		Message: fmt.Sprintf("Unsubscribed from `%s`", b.Request.Topic()),
	}

	// -- Critical section of code begins
	//
	// Attempting to write to a network resource,
	// shared among multiple go routines
	b.Lock.Lock()
	defer b.Lock.Unlock()

	if err := b.Connection.WriteJSON(resp); err != nil {
		log.Printf("[!] Failed to deliver `%s` unsubscription confirmation to client : %s\n", b.Request.Topic(), err.Error())
	}

}
