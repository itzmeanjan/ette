package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"

	d "github.com/itzmeanjan/ette/app/data"
	"github.com/itzmeanjan/ette/app/db"
)

// BlockConsumer - To be subscribed to `block` topic using this consumer handle
// and client connected using websocket needs to be delivered this piece of data
type BlockConsumer struct {
	Client      *redis.Client
	Request     *SubscriptionRequest
	UserAddress common.Address
	Connection  *websocket.Conn
	PubSub      *redis.PubSub
	DB          *gorm.DB
	Lock        *sync.Mutex
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

		if err := b.Connection.WriteJSON(&SubscriptionResponse{
			Code:    0,
			Message: "Bad API Key",
		}); err != nil {
			log.Printf("[!] Failed to deliver bad API key message to client : %s\n", err.Error())
		}

		return false

	}

	if !user.Enabled {

		if err := b.Connection.WriteJSON(&SubscriptionResponse{
			Code:    0,
			Message: "Bad API Key",
		}); err != nil {
			log.Printf("[!] Failed to deliver bad API key message to client : %s\n", err.Error())
		}

		return false

	}

	// Don't deliver data & close underlying connection
	// if client has crossed it's allowed data delivery limit
	if !db.IsUnderRateLimit(b.DB, b.UserAddress.Hex()) {

		if err := b.Connection.WriteJSON(&SubscriptionResponse{
			Code:    0,
			Message: "Crossed Allowed Rate Limit",
		}); err != nil {
			log.Printf("[!] Failed to deliver rate limit crossed message to client : %s\n", err.Error())
		}

		return false

	}

	var block d.Block

	_msg := []byte(msg)

	err := json.Unmarshal(_msg, &block)
	if err != nil {
		log.Printf("[!] Failed to decode published block data to JSON : %s\n", err.Error())
		return true
	}

	if b.SendData(&block) {
		db.PutDataDeliveryInfo(b.DB, b.UserAddress.Hex(), "/v1/ws/block", uint64(len(msg)))
		return true
	}

	return false

}

// SendData - Sending message to client application, connected over websocket
//
// If failed, we're going to remove subscription & close websocket
// connection ( connection might be already closed though )
func (b *BlockConsumer) SendData(data interface{}) bool {

	if err := b.Connection.WriteJSON(data); err != nil {
		log.Printf("[!] Failed to deliver `block` data to client : %s\n", err.Error())
		return false
	}

	log.Printf("[+] Delivered `block` data to client\n")
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

	if err := b.Connection.WriteJSON(resp); err != nil {
		log.Printf("[!] Failed to deliver `%s` unsubscription confirmation to client : %s\n", b.Request.Topic(), err.Error())
	}

}
