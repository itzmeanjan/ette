package data

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/adjust/rmq/v3"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"github.com/lib/pq"
)

// SyncState - Whether `ette` is synced with blockchain or not
type SyncState struct {
	Synced bool
}

// Block - Block related info to be delivered to client in this format
type Block struct {
	Hash       string `json:"hash" gorm:"column:hash"`
	Number     uint64 `json:"number" gorm:"column:number"`
	Time       uint64 `json:"time" gorm:"column:time"`
	ParentHash string `json:"parentHash" gorm:"column:parenthash"`
	Difficulty string `json:"difficulty" gorm:"column:difficulty"`
	GasUsed    uint64 `json:"gasUsed" gorm:"column:gasused"`
	GasLimit   uint64 `json:"gasLimit" gorm:"column:gaslimit"`
	Nonce      uint64 `json:"nonce" gorm:"column:nonce"`
}

// ToJSON - Encodes into JSON, to be supplied when queried for block data
func (b *Block) ToJSON() []byte {
	data, err := json.Marshal(b)
	if err != nil {
		log.Printf("[!] Failed to encode block data to JSON : %s\n", err.Error())
		return nil
	}

	return data
}

// Blocks - A set of blocks to be held, extracted from DB query result
// also to be supplied to client in JSON encoded form
type Blocks struct {
	Blocks []Block `json:"blocks"`
}

// ToJSON - Encoding into JSON, to be invoked when delivering query result to client
func (b *Blocks) ToJSON() []byte {
	data, err := json.Marshal(b)
	if err != nil {
		log.Printf("[!] Failed to encode block data to JSON : %s\n", err.Error())
		return nil
	}

	return data
}

// Transaction - Transaction holder struct, to be supplied when queried using tx hash
type Transaction struct {
	Hash      string `json:"hash" gorm:"column:hash"`
	From      string `json:"from" gorm:"column:from"`
	To        string `json:"to" gorm:"column:to"`
	Contract  string `json:"contract" gorm:"column:contract"`
	Gas       uint64 `json:"gas" gorm:"column:gas"`
	GasPrice  string `json:"gasPrice" gorm:"column:gasprice"`
	Cost      string `json:"cost" gorm:"column:cost"`
	Nonce     uint64 `json:"nonce" gorm:"column:nonce"`
	State     uint64 `json:"state" gorm:"column:state"`
	BlockHash string `json:"blockHash" gorm:"column:blockhash"`
}

// ToJSON - JSON encoder, to be invoked before delivering tx query data to client
func (t *Transaction) ToJSON() []byte {
	// When tx doesn't create contract i.e. normal tx
	if !strings.HasPrefix(t.Contract, "0x") {
		t.Contract = ""
	}

	// When tx creates contract
	if !strings.HasPrefix(t.To, "0x") {
		t.To = ""
	}

	data, err := json.Marshal(t)
	if err != nil {
		log.Printf("[!] Failed to encode transaction data to JSON : %s\n", err.Error())
		return nil
	}

	return data
}

// Transactions - Multiple transactions holder struct
type Transactions struct {
	Transactions []*Transaction `json:"transactions"`
}

// ToJSON - Encoding into JSON, to be invoked when delivering to client
func (t *Transactions) ToJSON() []byte {

	// Replacing contract address/ to address of tx
	// using empty string, if tx is normal tx/ creates contract
	// respectively
	//
	// We'll save some data transfer burden
	for _, v := range t.Transactions {
		if !strings.HasPrefix(v.Contract, "0x") {
			v.Contract = ""
		}

		// When tx creates contract
		if !strings.HasPrefix(v.To, "0x") {
			v.To = ""
		}
	}

	data, err := json.Marshal(t)
	if err != nil {
		log.Printf("[!] Failed to encode transaction data to JSON : %s\n", err.Error())
		return nil
	}

	return data
}

// Event - Single event entity holder, extracted from db
type Event struct {
	Origin          string         `gorm:"column:origin"`
	Index           uint           `gorm:"column:index"`
	Topics          pq.StringArray `gorm:"column:topics;type:text[]"`
	Data            []byte         `gorm:"column:data"`
	TransactionHash string         `gorm:"column:txhash"`
	BlockHash       string         `gorm:"column:blockhash"`
}

// MarshalJSON - Custom JSON encoder
func (e *Event) MarshalJSON() ([]byte, error) {

	data := ""
	if hex.EncodeToString(e.Data) != "" {
		data = fmt.Sprintf("0x%s", hex.EncodeToString(e.Data))
	}

	return []byte(fmt.Sprintf(`{"origin":%q,"index":%d,"topics":%v,"data":%q,"txHash":%q,"blockHash":%q}`,
		e.Origin,
		e.Index,
		strings.Join(
			strings.Fields(
				fmt.Sprintf("%q", e.Topics)), ","),
		data, e.TransactionHash, e.BlockHash)), nil

}

// ToJSON - Encoding into JSON
func (e *Event) ToJSON() []byte {

	data, err := json.Marshal(e)
	if err != nil {
		log.Printf("[!] Failed to encode event to JSON : %s\n", err.Error())
		return nil
	}

	return data

}

// Events - A collection of event holder, to be delivered to client in this form
type Events struct {
	Events []*Event `json:"events"`
}

// ToJSON - Encoding to JSON
func (e *Events) ToJSON() []byte {

	data, err := json.Marshal(e)
	if err != nil {
		log.Printf("[!] Failed to encode events to JSON : %s\n", err.Error())
		return nil
	}

	return data

}

// SubscriptionRequest - Real time data subscription/ unsubscription request
// needs to be sent in this form, from client application
type SubscriptionRequest struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// IsValidTopic - Checks whether topic to which client application is trying to
// subscribe to is valid one or not
func (s *SubscriptionRequest) IsValidTopic() bool {
	pattern, err := regexp.Compile("^(block|(transaction(/(0x[a-zA-Z0-9]{40}|\\*)(/(0x[a-zA-Z0-9]{40}|\\*))?)?))$")
	if err != nil {
		log.Printf("[!] Failed to parse regex pattern : %s\n", err.Error())
		return false
	}

	if pattern.MatchString(s.Name) {

		matches := pattern.FindStringSubmatch(s.Name)
		if strings.HasPrefix(matches[0], "block") {
			return true
		} else if strings.HasPrefix(matches[0], "transaction") {
			return true
		}

	}

	return false
}

// Validate - Validates request from client for subscription/ unsubscription
func (s *SubscriptionRequest) Validate(topics map[string]bool) bool {

	// --- Closure definition
	// Given associative array & key in array, checks whether entry exists or not
	// If yes, also return entry's boolean value
	checkEntryInAssociativeArray := func(name string) bool {
		v, ok := topics[name]
		if !ok {
			return false
		}

		return v
	}
	// ---

	var validated bool

	switch s.Type {
	case "subscribe":
		validated = s.IsValidTopic() && !checkEntryInAssociativeArray(s.Name)
	case "unsubscribe":
		validated = s.IsValidTopic() && checkEntryInAssociativeArray(s.Name)
	default:
		validated = false
	}

	return validated

}

// SubscriptionResponse - Real time data subscription/ unsubscription request to be responded with
// in this form
type SubscriptionResponse struct {
	Code    uint   `json:"code"`
	Message string `json:"msg"`
}

// BlockConsumer - Block data consumer to keep websocket connection handle
// so when data is delivered, it can let client application know about it
type BlockConsumer struct {
	Connections map[*websocket.Conn]bool
}

// Consume - When data is available on subscribed channel, consumer
// to be notified by invoking this method, where we send block data to client
// application ( which actually subscribed to  this channel )
// over websocket connection
func (b *BlockConsumer) Consume(delivery rmq.Delivery) {
	var block Block

	if err := json.Unmarshal([]byte(delivery.Payload()), &block); err != nil {

		log.Printf("[!] Bad delivery in block consumer : %s\n", err.Error())
		if err := delivery.Reject(); err != nil {
			log.Printf("[!] Failed to reject delivery from block consumer : %s\n", err.Error())
		}

		return
	}

	// bufferring connections to be removed, which are either unreachable
	// or unsubscribe from blocks topic
	//
	// removed connections to be considered as they've unsubscribed from
	// topic, so we won't attempt sending them block mining notification
	closedConnections := make([]*websocket.Conn, 0)

	for k, v := range b.Connections {
		if v {
			if err := k.WriteJSON(&block); err != nil {
				closedConnections = append(closedConnections, k)
			}
			continue
		}

		closedConnections = append(closedConnections, k)
	}

	// iterating over all those connections to be removed
	// and deleting them from map
	for _, v := range closedConnections {
		delete(b.Connections, v)
	}

	if err := delivery.Ack(); err != nil {
		log.Printf("[!] Failed to acknowledge delivery for block : %s\n", err.Error())
	}
}

// TransactionConsumer - Transaction topic subscribers websocket handles, to be
// used when letting them know about occurrence of their topic of interest
type TransactionConsumer struct {
	Connections map[*websocket.Conn]*SubscriptionRequest
}

// Consume - Whenever new transaction is found, this call back to be invoked by redis
// and deliver data, which can be used for finally letting client application
// connected via websocket inform regarding occurrence of their topic of interest
func (t *TransactionConsumer) Consume(delivery rmq.Delivery) {

	var transaction Transaction

	if err := json.Unmarshal([]byte(delivery.Payload()), &transaction); err != nil {

		log.Printf("[!] Bad delivery in transaction consumer : %s\n", err.Error())
		if err := delivery.Reject(); err != nil {
			log.Printf("[!] Failed to reject delivery from transaction consumer : %s\n", err.Error())
		}

		return

	}

	log.Printf("[+] %v\n", transaction)

	if err := delivery.Ack(); err != nil {
		log.Printf("[!] Failed to acknowledge delivery from transaction consumer : %s\n", err.Error())
	}

}

// PubSubClient - Handle for talking to redis for managing pub-sub
type PubSubClient struct {
	Client *redis.Client
}
