package data

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	cfg "github.com/itzmeanjan/ette/app/config"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/lib/pq"
)

// SyncState - Whether `ette` is synced with blockchain or not
type SyncState struct {
	Done      uint64
	Target    uint64
	StartedAt time.Time
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

// MarshalBinary - Implementing binary marshalling function, to be invoked
// by redis before publishing data on channel
func (t *Transaction) MarshalBinary() ([]byte, error) {
	return json.Marshal(t)
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

// AuthPayload - Payload to be sent in post request body, when performing
// login
type AuthPayload struct {
	Message   AuthPayloadMessage `json:"message" binding:"required"`
	Signature string             `json:"signature" binding:"required"`
}

// VerifySignature - Given recovered signer address from authentication payload
// checks whether person claiming to sign message has really signed or not
func (a *AuthPayload) VerifySignature(signer []byte) bool {
	if signer == nil {
		return false
	}

	return common.BytesToAddress(signer) == a.Message.Address
}

// IsAdmin - Given recovered signer address from authentication payload
// checks whether this address matches with admin address present in `.env` file
// or not
func (a *AuthPayload) IsAdmin(signer []byte) bool {
	if signer == nil {
		return false
	}

	return common.BytesToAddress(signer) == common.HexToAddress(cfg.Get("Admin"))
}

// HasExpired - Checking if message was signed with in
// `window` seconds ( will be kept generally 30s ) time span
// from current server time or not
func (a *AuthPayload) HasExpired(window int64) bool {
	return !(int64(a.Message.TimeStamp)+window >= time.Now().Unix())
}

// RecoverSigner - Given signed message & original message
// it recovers signer address as byte array, which is to be
// later used for matching against claimed signer address
func (a *AuthPayload) RecoverSigner() []byte {

	data := a.Message.ToJSON()
	if data == nil {
		return nil
	}

	signature, err := hexutil.Decode(a.Signature)
	if err != nil {
		return nil
	}

	if !(signature[64] == 27 || signature[64] == 28) {
		return nil
	}

	signature[64] -= 27

	pubKey, err := crypto.SigToPub(
		// After `Ethereum Signed Message` is prepended
		// we're performing keccak256 hash, which is actual message, signed in metamask
		crypto.Keccak256(
			// this is required, because for web3.personal.sign, it'll prepend this part
			// so we're also prepending it before recovering signature
			[]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(data), data))),
		signature)
	if err != nil {
		return nil
	}

	return crypto.PubkeyToAddress(*pubKey).Bytes()

}

// AuthPayloadMessage - Message to be signed by user
type AuthPayloadMessage struct {
	Address   common.Address `json:"address" binding:"required"`
	TimeStamp uint64         `json:"timestamp" binding:"required"`
}

// ToJSON - Encoding message to JSON, this is what was signed by user
func (a *AuthPayloadMessage) ToJSON() []byte {

	if data, err := json.Marshal(a); err == nil {
		return data
	}

	return nil
}
