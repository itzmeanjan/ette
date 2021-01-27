package pubsub

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	d "github.com/itzmeanjan/ette/app/data"
	"github.com/itzmeanjan/ette/app/db"
	"gorm.io/gorm"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"github.com/lib/pq"
)

// EventConsumer - Event consumption to be managed by this struct, when new websocket
// connection requests for receiving event data, it'll create this struct, with necessary pieces
// of information, which is to be required when delivering data & checking whether this connection
// has really requested notification for this event or not
type EventConsumer struct {
	Client      *redis.Client
	Request     *SubscriptionRequest
	UserAddress common.Address
	Connection  *websocket.Conn
	PubSub      *redis.PubSub
	DB          *gorm.DB
	Lock        *sync.Mutex
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

	// Before leaving this execution context, attempting to
	// unsubscribe client from `event` pubsub topic
	//
	// One attempt to unsubscribe gracefully
	defer e.Unsubscribe()

	for {

		// Checking if client is still subscribed to this topic
		// or not
		//
		// If not, we're cancelling this subscription
		if e.Request.Type == "unsubscribe" {
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

	user := db.GetUserFromAPIKey(e.DB, e.Request.APIKey)
	if user == nil {

		// -- Critical section of code begins
		//
		// Attempting to write to a network resource,
		// shared among multiple go routines
		e.Lock.Lock()

		if err := e.Connection.WriteJSON(&SubscriptionResponse{
			Code:    0,
			Message: "Bad API Key",
		}); err != nil {
			log.Printf("[!] Failed to deliver bad API key message to client : %s\n", err.Error())
		}

		e.Lock.Unlock()
		// -- ends here
		return false

	}

	if !user.Enabled {

		// -- Critical section of code begins
		//
		// Attempting to write to a network resource,
		// shared among multiple go routines
		e.Lock.Lock()

		if err := e.Connection.WriteJSON(&SubscriptionResponse{
			Code:    0,
			Message: "Bad API Key",
		}); err != nil {
			log.Printf("[!] Failed to deliver bad API key message to client : %s\n", err.Error())
		}

		e.Lock.Unlock()
		// -- ends here
		return false

	}

	// Don't deliver data & close underlying connection
	// if client has crossed it's allowed data delivery limit
	if !db.IsUnderRateLimit(e.DB, e.UserAddress.Hex()) {

		// -- Critical section of code begins
		//
		// Attempting to write to a network resource,
		// shared among multiple go routines
		e.Lock.Lock()

		if err := e.Connection.WriteJSON(&SubscriptionResponse{
			Code:    0,
			Message: "Crossed Allowed Rate Limit",
		}); err != nil {
			log.Printf("[!] Failed to deliver rate limit crossed message to client : %s\n", err.Error())
		}

		e.Lock.Unlock()
		// -- ends here
		return false

	}

	var event struct {
		Origin          string         `json:"origin"`
		Index           uint           `json:"index"`
		Topics          pq.StringArray `json:"topics"`
		Data            string         `json:"data"`
		TransactionHash string         `json:"txHash"`
		BlockHash       string         `json:"blockHash"`
	}

	_msg := []byte(msg)

	if err := json.Unmarshal(_msg, &event); err != nil {
		log.Printf("[!] Failed to decode published event data to JSON : %s\n", err.Error())
		return true
	}

	data := make([]byte, 0)
	var err error

	if len(event.Data) != 0 {
		data, err = hex.DecodeString(event.Data[2:])
	}

	if err != nil {
		log.Printf("[!] Failed to decode data field of event : %s\n", err.Error())
		return true
	}

	// If doesn't match with specified criteria, simply ignoring received data
	if !e.Request.DoesMatchWithPublishedEventData(&d.Event{
		Origin:          event.Origin,
		Index:           event.Index,
		Topics:          event.Topics,
		Data:            data,
		TransactionHash: event.TransactionHash,
		BlockHash:       event.BlockHash,
	}) {
		return true
	}

	if e.SendData(&event) {
		db.PutDataDeliveryInfo(e.DB, e.UserAddress.Hex(), "/v1/ws/event", uint64(len(msg)))
		return true
	}

	return false
}

// SendData - Sending message to client application, connected over websocket
//
// If failed, we're going to remove subscription & close websocket
// connection ( connection might be already closed though )
func (e *EventConsumer) SendData(data interface{}) bool {

	// -- Critical section of code begins
	//
	// Attempting to write to a network resource,
	// shared among multiple go routines
	e.Lock.Lock()
	defer e.Lock.Unlock()

	if err := e.Connection.WriteJSON(data); err != nil {
		log.Printf("[!] Failed to deliver `event` data to client : %s\n", err.Error())
		return false
	}

	return true

}

// Unsubscribe - Unsubscribe from event data publishing topic, to be called
// when stopping to listen data being published on this pubsub channel
// due to client has requested a unsubscription/ network connection got hampered
func (e *EventConsumer) Unsubscribe() {

	if e.PubSub == nil {
		log.Printf("[!] Bad attempt to unsubscribe from `%s` topic\n", e.Request.Topic())
		return
	}

	if err := e.PubSub.Unsubscribe(context.Background(), e.Request.Topic()); err != nil {
		log.Printf("[!] Failed to unsubscribe from `%s` topic : %s\n", e.Request.Topic(), err.Error())
		return
	}

	resp := &SubscriptionResponse{
		Code:    1,
		Message: fmt.Sprintf("Unsubscribed from `%s`", e.Request.Topic()),
	}

	// -- Critical section of code begins
	//
	// Attempting to write to a network resource,
	// shared among multiple go routines
	e.Lock.Lock()
	defer e.Lock.Unlock()

	if err := e.Connection.WriteJSON(resp); err != nil {
		log.Printf("[!] Failed to deliver `%s` unsubscription confirmation to client : %s\n", e.Request.Topic(), err.Error())
	}

}
