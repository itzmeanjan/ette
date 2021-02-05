package pubsub

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	d "github.com/itzmeanjan/ette/app/data"
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
	Lock       *sync.Mutex
}

// Subscribe - Subscribe to `transaction` topic, under which all transaction related data to be published
func (t *TransactionConsumer) Subscribe() {
	t.PubSub = t.Client.Subscribe(context.Background(), t.Request.Topic())
}

// Listen - Listener function, which keeps looping in infinite loop
// and reads data from subcribed channel, which also gets delivered to client application
func (t *TransactionConsumer) Listen() {

	// When leaving this execution scope, attempting to unsubscribe
	// client from pubsub topic, it was listening to, in a graceful fashion
	defer t.Unsubscribe()

	for {

		// Checking if client is still subscribed to this topic
		// or not
		//
		// If not, we're cancelling this subscription
		if t.Request.Type == "unsubscribe" {
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

	user := db.GetUserFromAPIKey(t.DB, t.Request.APIKey)
	if user == nil {

		// -- Critical section of code begins
		//
		// Attempting to write to a network resource,
		// shared among multiple go routines
		t.Lock.Lock()

		if err := t.Connection.WriteJSON(&SubscriptionResponse{
			Code:    0,
			Message: "Bad API Key",
		}); err != nil {
			log.Printf("[!] Failed to deliver bad API key message to client : %s\n", err.Error())
		}

		t.Lock.Unlock()
		// -- ends here
		return false

	}

	if !user.Enabled {

		// -- Critical section of code begins
		//
		// Attempting to write to a network resource,
		// shared among multiple go routines
		t.Lock.Lock()

		if err := t.Connection.WriteJSON(&SubscriptionResponse{
			Code:    0,
			Message: "Bad API Key",
		}); err != nil {
			log.Printf("[!] Failed to deliver bad API key message to client : %s\n", err.Error())
		}

		t.Lock.Unlock()
		// -- ends here
		return false

	}

	// Don't deliver data & close underlying connection
	// if client has crossed it's allowed data delivery limit
	if !db.IsUnderRateLimit(t.DB, user.Address) {

		// -- Critical section of code begins
		//
		// Attempting to write to a network resource,
		// shared among multiple go routines
		t.Lock.Lock()

		if err := t.Connection.WriteJSON(&SubscriptionResponse{
			Code:    0,
			Message: "Crossed Allowed Rate Limit",
		}); err != nil {
			log.Printf("[!] Failed to deliver rate limit crossed message to client : %s\n", err.Error())
		}

		t.Lock.Unlock()
		// -- ends here
		return false

	}

	// Creating this temporary struct definition here, because
	// while unmarshalling JSON it was failing in `{ Data: []byte }`
	// part, because it was byte array
	//
	// Now as it's first decoded as string, then it'll converted to byte array
	// if it's not a empty string
	var transaction struct {
		Hash      string `json:"hash"`
		From      string `json:"from"`
		To        string `json:"to"`
		Contract  string `json:"contract"`
		Value     string `json:"value"`
		Data      string `json:"data"`
		Gas       uint64 `json:"gas"`
		GasPrice  string `json:"gasPrice"`
		Cost      string `json:"cost"`
		Nonce     uint64 `json:"nonce"`
		State     uint64 `json:"state"`
		BlockHash string `json:"blockHash"`
	}

	_msg := []byte(msg)

	if err := json.Unmarshal(_msg, &transaction); err != nil {
		log.Printf("[!] Failed to decode published transaction data to JSON : %s\n", err.Error())
		return true
	}

	data := make([]byte, 0)
	var err error

	// If `data` field is not empty, we'll try to decode
	// part to tx data, after slicing out `0x` part prepended
	// to it
	if len(transaction.Data) != 0 {
		data, err = hex.DecodeString(transaction.Data[2:])
	}

	if err != nil {
		log.Printf("[!] Failed to decode data field of transaction : %s\n", err.Error())
		return true
	}

	// If doesn't match, simply ignoring received data
	if !t.Request.DoesMatchWithPublishedTransactionData(&d.Transaction{
		Hash:      transaction.Hash,
		From:      transaction.From,
		To:        transaction.To,
		Contract:  transaction.Contract,
		Value:     transaction.Value,
		Data:      data,
		Gas:       transaction.Gas,
		GasPrice:  transaction.GasPrice,
		Cost:      transaction.Cost,
		Nonce:     transaction.Nonce,
		State:     transaction.State,
		BlockHash: transaction.BlockHash,
	}) {
		return true
	}

	if t.SendData(&transaction) {
		db.PutDataDeliveryInfo(t.DB, user.Address, "/v1/ws/transaction", uint64(len(msg)))
		return true
	}

	return false
}

// SendData - Sending message to client application, connected over websocket
//
// If failed, we're going to remove subscription & close websocket
// connection ( connection might be already closed though )
func (t *TransactionConsumer) SendData(data interface{}) bool {

	// -- Critical section of code begins
	//
	// Attempting to write to a network resource,
	// shared among multiple go routines
	t.Lock.Lock()
	defer t.Lock.Unlock()

	if err := t.Connection.WriteJSON(data); err != nil {
		log.Printf("[!] Failed to deliver `transaction` data to client : %s\n", err.Error())
		return false
	}

	return true

}

// Unsubscribe - Unsubscribe from transactions pubsub topic, which client has subscribed to
func (t *TransactionConsumer) Unsubscribe() {

	if t.PubSub == nil {
		log.Printf("[!] Bad attempt to unsubscribe from `%s` topic\n", t.Request.Topic())
		return
	}

	if err := t.PubSub.Unsubscribe(context.Background(), t.Request.Topic()); err != nil {
		log.Printf("[!] Failed to unsubscribe from `%s` topic : %s\n", t.Request.Topic(), err.Error())
		return
	}

	resp := &SubscriptionResponse{
		Code:    1,
		Message: fmt.Sprintf("Unsubscribed from `%s`", t.Request.Topic()),
	}

	// -- Critical section of code begins
	//
	// Attempting to write to a network resource,
	// shared among multiple go routines
	t.Lock.Lock()
	defer t.Lock.Unlock()

	if err := t.Connection.WriteJSON(resp); err != nil {
		log.Printf("[!] Failed to deliver `%s` unsubscription confirmation to client : %s\n", t.Request.Topic(), err.Error())
	}

}
