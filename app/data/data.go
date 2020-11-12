package data

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"

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

// BlockConsumer - To be subscribed to `block` topic using this consumer handle
// and client connected using websocket needs to be delivered this piece of data
type BlockConsumer struct {
	Client     *redis.Client
	Channel    string
	Connection *websocket.Conn
	PubSub     *redis.PubSub
}

// Subscribe - ...
func (b *BlockConsumer) Subscribe() {
	b.PubSub = b.Client.Subscribe(context.Background(), b.Channel)
}
