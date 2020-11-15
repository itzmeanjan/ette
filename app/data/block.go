package data

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
)

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
	b.PubSub = b.Client.Subscribe(context.Background(), b.Request.Topic())
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

			if err := b.Connection.WriteJSON(&SubscriptionResponse{
				Code:    1,
				Message: "Unsubscribed from `block`",
			}); err != nil {
				log.Printf("[!] Failed to deliver block unsubscription confirmation to client : %s\n", err.Error())
			}

			if err := b.PubSub.Unsubscribe(context.Background(), b.Request.Topic()); err != nil {
				log.Printf("[!] Failed to unsubscribe from `block` topic : %s\n", err.Error())
			}
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
	var block Block

	_msg := []byte(msg)

	err := json.Unmarshal(_msg, &block)
	if err != nil {
		log.Printf("[!] Failed to decode published block data to JSON : %s\n", err.Error())
		return true
	}

	return b.SendData(&block)
}

// SendData - Sending message to client application, connected over websocket
//
// If failed, we're going to remove subscription & close websocket
// connection ( connection might be already closed though )
func (b *BlockConsumer) SendData(data interface{}) bool {
	if err := b.Connection.WriteJSON(data); err != nil {
		log.Printf("[!] Failed to deliver `block` data to client : %s\n", err.Error())

		if err = b.PubSub.Unsubscribe(context.Background(), b.Request.Topic()); err != nil {
			log.Printf("[!] Failed to unsubscribe from `block` topic : %s\n", err.Error())
		}

		if err = b.Connection.Close(); err != nil {
			log.Printf("[!] Failed to close websocket connection : %s\n", err.Error())
		}

		return false
	}

	log.Printf("[!] Delivered `block` data to client\n")
	return true
}
