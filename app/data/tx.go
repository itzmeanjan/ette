package data

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

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
