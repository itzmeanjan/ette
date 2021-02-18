package pubsub

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"github.com/itzmeanjan/ette/app/data"
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
	Requests   map[string]*SubscriptionRequest
	Connection *websocket.Conn
	PubSub     *redis.PubSub
	DB         *gorm.DB
	ConnLock   *sync.Mutex
	TopicLock  *sync.RWMutex
	Counter    *data.SendReceiveCounter
}

// Subscribe - Subscribe to `transaction` topic, under which all transaction related data to be published
func (t *TransactionConsumer) Subscribe() {
	t.PubSub = t.Client.Subscribe(context.Background(), "transaction")
}

// Listen - Listener function, which keeps looping in infinite loop
// and reads data from subcribed channel, which also gets delivered to client application
func (t *TransactionConsumer) Listen() {

	for {

		msg, err := t.PubSub.ReceiveTimeout(context.Background(), time.Second)
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

			t.SendData(&SubscriptionResponse{
				Code:    1,
				Message: "Subscribed to `transaction`",
			})

		case *redis.Message:
			t.Send(m.Payload)

		}

	}

}

// Send - Tries to deliver subscribed transaction data to client application
// connected over websocket
func (t *TransactionConsumer) Send(msg string) {

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
		return
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
		return
	}

	tx := &d.Transaction{
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
	}

	var request *SubscriptionRequest

	// -- Shared memory being read from concurrently
	// running thread of execution, with lock
	t.TopicLock.RLock()

	for _, v := range t.Requests {

		if v.DoesMatchWithPublishedTransactionData(tx) {
			request = v
			break
		}

	}

	t.TopicLock.RUnlock()
	// -- Lock released, shared memory reading done

	// Can't proceed with this anymore, because failed to find
	// respective subscription request
	if request == nil {
		return
	}

	user := db.GetUserFromAPIKey(t.DB, request.APIKey)
	if user == nil {

		// -- Critical section of code begins
		//
		// Attempting to write to a network resource,
		// shared among multiple go routines
		t.ConnLock.Lock()

		if err := t.Connection.WriteJSON(&SubscriptionResponse{
			Code:    0,
			Message: "Bad API Key",
		}); err != nil {
			log.Printf("[!] Failed to deliver bad API key message to client : %s\n", err.Error())
		}

		t.ConnLock.Unlock()
		// -- ends here

		// Because we're writing to socket
		t.Counter.IncrementSend(1)
		return

	}

	if !user.Enabled {

		// -- Critical section of code begins
		//
		// Attempting to write to a network resource,
		// shared among multiple go routines
		t.ConnLock.Lock()

		if err := t.Connection.WriteJSON(&SubscriptionResponse{
			Code:    0,
			Message: "Bad API Key",
		}); err != nil {
			log.Printf("[!] Failed to deliver bad API key message to client : %s\n", err.Error())
		}

		t.ConnLock.Unlock()
		// -- ends here

		// Because we're writing to socket
		t.Counter.IncrementSend(1)
		return

	}

	// Don't deliver data & close underlying connection
	// if client has crossed it's allowed data delivery limit
	if !db.IsUnderRateLimit(t.DB, user.Address) {

		// -- Critical section of code begins
		//
		// Attempting to write to a network resource,
		// shared among multiple go routines
		t.ConnLock.Lock()

		if err := t.Connection.WriteJSON(&SubscriptionResponse{
			Code:    0,
			Message: "Crossed Allowed Rate Limit",
		}); err != nil {
			log.Printf("[!] Failed to deliver rate limit crossed message to client : %s\n", err.Error())
		}

		t.ConnLock.Unlock()
		// -- ends here

		// Because we're writing to socket
		t.Counter.IncrementSend(1)
		return

	}

	if t.SendData(&transaction) {
		db.PutDataDeliveryInfo(t.DB, user.Address, "/v1/ws/transaction", uint64(len(msg)))
	}

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
	t.ConnLock.Lock()
	defer t.ConnLock.Unlock()

	if err := t.Connection.WriteJSON(data); err != nil {
		log.Printf("[!] Failed to deliver `transaction` data to client : %s\n", err.Error())
		return false
	}

	// Because we're writing to socket
	t.Counter.IncrementSend(1)

	return true

}

// Unsubscribe - Unsubscribe from transactions pubsub topic, which client has subscribed to
func (t *TransactionConsumer) Unsubscribe() {

	if t.PubSub == nil {
		log.Printf("[!] Bad attempt to unsubscribe from `transaction` topic\n")
		return
	}

	if err := t.PubSub.Unsubscribe(context.Background(), "transaction"); err != nil {
		log.Printf("[!] Failed to unsubscribe from `transaction` topic : %s\n", err.Error())
		return
	}

	resp := &SubscriptionResponse{
		Code:    1,
		Message: "Unsubscribed from `transaction`",
	}

	// -- Critical section of code begins
	//
	// Attempting to write to a network resource,
	// shared among multiple go routines
	t.ConnLock.Lock()
	defer t.ConnLock.Unlock()

	if err := t.Connection.WriteJSON(resp); err != nil {

		log.Printf("[!] Failed to deliver `transaction` unsubscription confirmation to client : %s\n", err.Error())
		return

	}

	// Because we're writing to socket
	t.Counter.IncrementSend(1)

}
