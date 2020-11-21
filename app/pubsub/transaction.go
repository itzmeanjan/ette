package pubsub

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"github.com/itzmeanjan/ette/app/data"
	"github.com/itzmeanjan/ette/app/db"
	"gorm.io/gorm"
)

// TransactionConsumer - Transaction consumer info holder struct, to be used
// for handling reception of published data & checking whether this client has really
// subscribed for this data or not
//
// If yes, also deliver data to client application, connected over websocket
type TransactionConsumer struct {
	Client     *redis.Client
	Request    *SubscriptionRequest
	Connection *websocket.Conn
	PubSub     *redis.PubSub
	DB         *gorm.DB
}

// Subscribe - Subscribe to `transaction` topic, under which all transaction related data to be published
func (t *TransactionConsumer) Subscribe() {
	t.PubSub = t.Client.Subscribe(context.Background(), t.Request.Topic())
}

// Listen - Listener function, which keeps looping in infinite loop
// and reads data from subcribed channel, which also gets delivered to client application
func (t *TransactionConsumer) Listen() {

	for {

		// Checking if client is still subscribed to this topic
		// or not
		//
		// If not, we're cancelling this subscription
		if t.Request.Type == "unsubscribe" {

			if err := t.Connection.WriteJSON(&SubscriptionResponse{
				Code:    1,
				Message: "Unsubscribed from `transaction`",
			}); err != nil {
				log.Printf("[!] Failed to deliver transaction unsubscription confirmation to client : %s\n", err.Error())
			}

			if err := t.PubSub.Unsubscribe(context.Background(), t.Request.Topic()); err != nil {
				log.Printf("[!] Failed to unsubscribe from `transaction` topic : %s\n", err.Error())
			}
			break

		}

		msg, err := t.PubSub.ReceiveTimeout(context.Background(), time.Second)
		if err != nil {
			continue
		}

		// To be used for checking whether delivering data to client went successful or not
		status := true

		switch m := msg.(type) {
		case *redis.Subscription:
			status = t.SendData(&SubscriptionResponse{
				Code:    1,
				Message: "Subscribed to `transaction`",
			})
		case *redis.Message:
			status = t.Send(m.Payload)
		}

		if !status {
			break
		}
	}

}

// Send - Tries to deliver subscribed transaction data to client application
// connected over websocket
func (t *TransactionConsumer) Send(msg string) bool {

	// Don't deliver data & close underlying connection
	// if client has crossed it's allowed data delivery limit
	if !db.IsUnderRateLimit(t.DB, t.Request.APIKey) {
		return false
	}

	var transaction data.Transaction

	_msg := []byte(msg)

	if err := json.Unmarshal(_msg, &transaction); err != nil {
		log.Printf("[!] Failed to decode published transaction data to JSON : %s\n", err.Error())
		return true
	}

	// If doesn't match, simply ignoring received data
	if !t.Request.DoesMatchWithPublishedTransactionData(&transaction) {
		return true
	}

	if t.SendData(&transaction) {
		db.PutDataDeliveryInfo(t.DB, t.Request.APIKey, "/v1/ws/transaction", uint64(len(msg)))
		return true
	}

	return false
}

// SendData - Sending message to client application, connected over websocket
//
// If failed, we're going to remove subscription & close websocket
// connection ( connection might be already closed though )
func (t *TransactionConsumer) SendData(data interface{}) bool {
	if err := t.Connection.WriteJSON(data); err != nil {
		log.Printf("[!] Failed to deliver `transaction` data to client : %s\n", err.Error())

		if err = t.PubSub.Unsubscribe(context.Background(), t.Request.Topic()); err != nil {
			log.Printf("[!] Failed to unsubscribe from `transaction` topic : %s\n", err.Error())
		}

		if err = t.Connection.Close(); err != nil {
			log.Printf("[!] Failed to close websocket connection : %s\n", err.Error())
		}

		return false
	}

	log.Printf("[!] Delivered `transaction` data to client\n")
	return true
}
