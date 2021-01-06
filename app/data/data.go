package data

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-redis/redis/v8"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// RedisInfo - Holds redis related information in this struct, to be used
// when passing to functions as argument
type RedisInfo struct {
	Client    *redis.Client // using this object `ette` will talk to Redis
	QueueName string        // retry queue name, for storing block numbers
}

// ResultStatus - Keeps track of how many operations went successful
// and how many of them failed
type ResultStatus struct {
	Success uint64
	Failure uint64
}

// Total - Returns total count of operations which were supposed to be
// performed
//
// To be useful when deciding whether all go routines have sent their status i.e. completed
// their task or not
func (r ResultStatus) Total() uint64 {
	return r.Success + r.Failure
}

// Job - For running a block fetching job, these are all the information which are required
type Job struct {
	Client      *ethclient.Client
	DB          *gorm.DB
	RedisClient *redis.Client
	RedisKey    string
	Block       uint64
	Lock        *sync.Mutex
	Synced      *SyncState
}

// BlockChainNodeConnection - Holds network connection object for blockchain nodes
//
// Use `RPC` i.e. HTTP based connection, for querying blockchain for data
// Use `Websocket` for real-time listening of events in blockchain
type BlockChainNodeConnection struct {
	RPC       *ethclient.Client
	Websocket *ethclient.Client
}

// SyncState - Whether `ette` is synced with blockchain or not
type SyncState struct {
	Done                uint64
	StartedAt           time.Time
	BlockCountAtStartUp uint64
	NewBlocksInserted   uint64
}

// BlockCountInDB - Blocks currently present in database
func (s *SyncState) BlockCountInDB() uint64 {
	return s.BlockCountAtStartUp + s.NewBlocksInserted
}

// Block - Block related info to be delivered to client in this format
type Block struct {
	Hash                string  `json:"hash" gorm:"column:hash"`
	Number              uint64  `json:"number" gorm:"column:number"`
	Time                uint64  `json:"time" gorm:"column:time"`
	ParentHash          string  `json:"parentHash" gorm:"column:parenthash"`
	Difficulty          string  `json:"difficulty" gorm:"column:difficulty"`
	GasUsed             uint64  `json:"gasUsed" gorm:"column:gasused"`
	GasLimit            uint64  `json:"gasLimit" gorm:"column:gaslimit"`
	Nonce               uint64  `json:"nonce" gorm:"column:nonce"`
	Miner               string  `json:"miner" gorm:"column:miner"`
	Size                float64 `json:"size" gorm:"column:size"`
	TransactionRootHash string  `json:"txRootHash" gorm:"column:txroothash"`
	ReceiptRootHash     string  `json:"receiptRootHash" gorm:"column:receiptroothash"`
}

// MarshalBinary - Implementing binary marshalling function, to be invoked
// by redis before publishing data on channel
func (b *Block) MarshalBinary() ([]byte, error) {
	return json.Marshal(b)
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
	Blocks []*Block `json:"blocks"`
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
	Value     string `json:"value" gorm:"column:value"`
	Data      []byte `json:"data" gorm:"column:data"`
	Gas       uint64 `json:"gas" gorm:"column:gas"`
	GasPrice  string `json:"gasPrice" gorm:"column:gasprice"`
	Cost      string `json:"cost" gorm:"column:cost"`
	Nonce     uint64 `json:"nonce" gorm:"column:nonce"`
	State     uint64 `json:"state" gorm:"column:state"`
	BlockHash string `json:"blockHash" gorm:"column:blockhash"`
}

// MarshalBinary - Implementing binary marshalling function, to be invoked
// by redis before publishing data on channel
func (t *Transaction) MarshalBinary() ([]byte, error) {
	return json.Marshal(t)
}

// MarshalJSON - Custom JSON encoder
func (t *Transaction) MarshalJSON() ([]byte, error) {

	data := ""
	if _h := hex.EncodeToString(t.Data); _h != "" {
		data = fmt.Sprintf("0x%s", _h)
	}

	// When tx doesn't create contract i.e. normal tx
	if !strings.HasPrefix(t.Contract, "0x") {
		return []byte(fmt.Sprintf(`{"hash":%q,"from":%q,"to":%q,"value":%q,"data":%q,"gas":%d,"gasPrice":%q,"cost":%q,"nonce":%d,"state":%d,"blockHash":%q}`,
			t.Hash, t.From, t.To, t.Value,
			data, t.Gas, t.GasPrice, t.Cost, t.Nonce, t.State, t.BlockHash)), nil
	}

	// When tx creates contract
	return []byte(fmt.Sprintf(`{"hash":%q,"from":%q,"contract":%q,"value":%q,"data":%q,"gas":%d,"gasPrice":%q,"cost":%q,"nonce":%d,"state":%d,"blockHash":%q}`,
		t.Hash, t.From, t.Contract, t.Value,
		data, t.Gas, t.GasPrice, t.Cost, t.Nonce, t.State, t.BlockHash)), nil

}

// ToJSON - JSON encoder, to be invoked before delivering tx query data to client
func (t *Transaction) ToJSON() []byte {

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

	data, err := json.Marshal(t)
	if err != nil {
		log.Printf("[!] Failed to encode transactions data to JSON : %s\n", err.Error())
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

// MarshalBinary - Implementing binary marshalling function, to be invoked
// by redis before publishing data on channel
func (e *Event) MarshalBinary() ([]byte, error) {
	return e.MarshalJSON()
}

// MarshalJSON - Custom JSON encoder
func (e *Event) MarshalJSON() ([]byte, error) {

	data := ""
	if _h := hex.EncodeToString(e.Data); _h != "" && _h != strings.Repeat("0", 64) {
		data = fmt.Sprintf("0x%s", _h)
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
