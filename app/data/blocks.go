package data

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
)

// NewBlockConsumer - Creating one new block data consumer, which will subscribe to block
// topic & listen for data being published on this channel, which will eventually be
// delivered to client application over websocket connection
func NewBlockConsumer(client *redis.Client, conn *websocket.Conn, req *SubscriptionRequest) {
	consumer := BlockConsumer{
		Client:     client,
		Request:    req,
		Connection: conn,
	}

	consumer.Subscribe()
	go consumer.Listen()
}

// BlockConsumer - To be subscribed to `block` topic using this consumer handle
// and client connected using websocket needs to be delivered this piece of data
type BlockConsumer struct {
	Client     *redis.Client
	Request    *SubscriptionRequest
	Connection *websocket.Conn
	PubSub     *redis.PubSub
}

// Subscribe - Subscribe to `block` channel
func (b *BlockConsumer) Subscribe() {
	b.PubSub = b.Client.Subscribe(context.Background(), b.Request.Name)
}

// Listen - Listener function, which keeps looping in infinite loop
// and reads data from subcribed channel, which also gets delivered to client application
func (b *BlockConsumer) Listen() {

	for {

		// Checking if client is still subscribed to this topic
		// or not
		//
		// If not, we're cancelling this subscription
		if b.Request.Type == "unsubscribe" {

			if err := b.PubSub.Unsubscribe(context.Background(), b.Request.Name); err != nil {
				log.Printf("[!] Failed to unsubscribe from block event : %s\n", err.Error())
			}
			break

		}

		msg, err := b.PubSub.ReceiveTimeout(context.Background(), time.Second)
		if err != nil {
			continue
		}

		switch m := msg.(type) {
		case *redis.Subscription:
			if !b.SendConfirmation() {
				break
			}
		case *redis.Message:
			if !b.Send(m.Payload) {
				break
			}
		}
	}

}

// Send - Tries to deliver subscribed block data to client application
// connected over websocket
func (b *BlockConsumer) Send(msg string) bool {
	var block Block

	_msg := []byte(msg)

	err := json.Unmarshal(_msg, &block)
	if err != nil {
		log.Printf("[!] Failed to decode published block data to JSON : %s\n", err.Error())
		return true
	}

	if err = b.Connection.WriteJSON(&block); err != nil {
		log.Printf("[!] Failed to deliver block data to client : %s\n", err.Error())

		if err = b.PubSub.Unsubscribe(context.Background(), b.Request.Name); err != nil {
			log.Printf("[!] Failed to unsubscribe from block event : %s\n", err.Error())
		}

		if err = b.Connection.Close(); err != nil {
			log.Printf("[!] Failed to close websocket connection : %s\n", err.Error())
		}

		return false
	}

	log.Printf("[!] Delivered block data to client\n")
	return true
}

// SendConfirmation - Sending confirmation message i.e. block subscription has been confirmed
// for client. If unable to send it, cancels subscription & closes underlying websocket connection
//
// Websocket connection may already be closed, in that case it'll simply return
func (b *BlockConsumer) SendConfirmation() bool {

	if err := b.Connection.WriteJSON(&SubscriptionResponse{
		Code:    1,
		Message: "Subscribed to `block`",
	}); err != nil {
		log.Printf("[!] Failed to deliver block subscription confirmation to client : %s\n", err.Error())

		if err = b.PubSub.Unsubscribe(context.Background(), b.Request.Name); err != nil {
			log.Printf("[!] Failed to unsubscribe from block event : %s\n", err.Error())
		}

		if err = b.Connection.Close(); err != nil {
			log.Printf("[!] Failed to close websocket connection : %s\n", err.Error())
		}

		return false
	}

	log.Printf("[!] Delivered block subscription confirmation to client\n")
	return true
}
