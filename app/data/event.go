package data

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"github.com/lib/pq"
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
			status = e.SendData(&SubscriptionResponse{
				Code:    1,
				Message: "Subscribed to `event`",
			})
		case *redis.Message:
			status = e.Send(m.Payload)
		}

		if !status {
			break
		}
	}

}

// Send - Sending event occurrence data to client application, which has subscribed to this event
// & connected over websocket
func (e *EventConsumer) Send(msg string) bool {
	var event struct {
		Origin          string         `json:"origin"`
		Index           uint           `json:"index"`
		Topics          pq.StringArray `json:"topics"`
		Data            string         `json:"data"`
		TransactionHash string         `json:"txHash"`
		BlockHash       string         `json:"blockHash"`
	}

	_msg := []byte(msg)

	err := json.Unmarshal(_msg, &event)
	if err != nil {
		log.Printf("[!] Failed to decode published event data to JSON : %s\n", err.Error())
		return true
	}

	data, err := hex.DecodeString(event.Data[2:])
	if err != nil {
		log.Printf("[!] Failed to decode data field of event : %s\n", err.Error())
		return true
	}

	// If doesn't match with specified criteria, simply ignoring received data
	if !e.Request.DoesMatchWithPublishedEventData(&Event{
		Origin:          event.Origin,
		Index:           event.Index,
		Topics:          event.Topics,
		Data:            data,
		TransactionHash: event.TransactionHash,
		BlockHash:       event.BlockHash,
	}) {
		return true
	}

	return e.SendData(&event)
}

// SendData - Sending message to client application, connected over websocket
//
// If failed, we're going to remove subscription & close websocket
// connection ( connection might be already closed though )
func (e *EventConsumer) SendData(data interface{}) bool {
	if err := e.Connection.WriteJSON(data); err != nil {
		log.Printf("[!] Failed to deliver `event` data to client : %s\n", err.Error())

		if err = e.PubSub.Unsubscribe(context.Background(), e.Request.Topic()); err != nil {
			log.Printf("[!] Failed to unsubscribe from `event` topic : %s\n", err.Error())
		}

		if err = e.Connection.Close(); err != nil {
			log.Printf("[!] Failed to close websocket connection : %s\n", err.Error())
		}

		return false
	}

	log.Printf("[!] Delivered `event` data to client\n")
	return true
}
