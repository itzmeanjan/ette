package data

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
)

// EventConsumer - Event consumption to be managed by this struct, when new websocket
// connection requests for receiving event data, it'll create this struct, with necessary pieces
// of information, which is to be required when delivering data & checking whether this connection
// has really requested notification for this event or not
type EventConsumer struct {
	Client     *redis.Client
	Request    *SubscriptionRequest
	Connection *websocket.Conn
	PubSub     *redis.PubSub
}

// Subscribe - Event consumer is subscribing to `event` topic,
// where all event related data to be published
func (e *EventConsumer) Subscribe() {
	e.PubSub = e.Client.Subscribe(context.Background(), e.Request.Topic())
}

// Listen - Polling for new data published in `event` topic periodically
// and sending data to subscribed to client ( connected over websocket )
// if client has subscribed to get notified on occurrence of this event
func (e *EventConsumer) Listen() {

	for {

		// Checking if client is still subscribed to this topic
		// or not
		//
		// If not, we're cancelling this subscription
		if e.Request.Type == "unsubscribe" {

			if err := e.Connection.WriteJSON(&SubscriptionResponse{
				Code:    1,
				Message: "Unsubscribed from `event`",
			}); err != nil {
				log.Printf("[!] Failed to deliver event unsubscription confirmation to client : %s\n", err.Error())
			}

			if err := e.PubSub.Unsubscribe(context.Background(), e.Request.Topic()); err != nil {
				log.Printf("[!] Failed to unsubscribe from `event` topic : %s\n", err.Error())
			}
			break

		}

		msg, err := e.PubSub.ReceiveTimeout(context.Background(), time.Second)
		if err != nil {
			continue
		}

		// To be used for checking whether delivering data to client went successful or not
		status := true

		switch m := msg.(type) {
		case *redis.Subscription:
			status = e.SendConfirmation()
		case *redis.Message:
			status = e.Send(m.Payload)
		}

		if !status {
			break
		}
	}

}

// Send - Sending event occurrence data to client application, which has subscribed to event
// & connected over websocket
//
// If failed, we're going to safely assume connection is closed, so subscription is also removed
func (e *EventConsumer) Send(msg string) bool {
	var block Block

	_msg := []byte(msg)

	err := json.Unmarshal(_msg, &block)
	if err != nil {
		log.Printf("[!] Failed to decode published event data to JSON : %s\n", err.Error())
		return true
	}

	if err = e.Connection.WriteJSON(&block); err != nil {
		log.Printf("[!] Failed to deliver event data to client : %s\n", err.Error())

		if err = e.PubSub.Unsubscribe(context.Background(), e.Request.Topic()); err != nil {
			log.Printf("[!] Failed to unsubscribe from `event` topic : %s\n", err.Error())
		}

		if err = e.Connection.Close(); err != nil {
			log.Printf("[!] Failed to close websocket connection : %s\n", err.Error())
		}

		return false
	}

	log.Printf("[!] Delivered event data to client\n")
	return true
}

// SendConfirmation - Sending confirmation message to client application,
// to denote subscription to `event` topic, has been done successfully & whenever
// new data gets published in this channel & client is accessible, it'll be
// delivered to client
func (e *EventConsumer) SendConfirmation() bool {

	if err := e.Connection.WriteJSON(&SubscriptionResponse{
		Code:    1,
		Message: "Subscribed to `event`",
	}); err != nil {
		log.Printf("[!] Failed to deliver event subscription confirmation to client : %s\n", err.Error())

		if err = e.PubSub.Unsubscribe(context.Background(), e.Request.Topic()); err != nil {
			log.Printf("[!] Failed to unsubscribe from `event` topic : %s\n", err.Error())
		}

		if err = e.Connection.Close(); err != nil {
			log.Printf("[!] Failed to close websocket connection : %s\n", err.Error())
		}

		return false
	}

	log.Printf("[!] Delivered event subscription confirmation to client\n")
	return true
}
