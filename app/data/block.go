package data

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
)

// Block - Block related info to be delivered to client in this format
type Block struct {
	Hash                string  `json:"hash" gorm:"column:hash"`
	Number              uint64  `json:"number" gorm:"column:number"`
	Time                uint64  `json:"time" gorm:"column:time"`
	ParentHash          string  `json:"parentHash" gorm:"column:parenthash"`
	Difficulty          string  `json:"difficulty" gorm:"column:difficulty"`
	GasUsed             uint64  `json:"gasUsed" gorm:"column:gasused"`
	GasLimit            uint64  `json:"gasLimit" gorm:"column:gaslimit"`
	Nonce               string  `json:"nonce" gorm:"column:nonce"`
	Miner               string  `json:"miner" gorm:"column:miner"`
	Size                float64 `json:"size" gorm:"column:size"`
	StateRootHash       string  `json:"stateRootHash" gorm:"column:stateroothash"`
	UncleHash           string  `json:"uncleHash" gorm:"column:unclehash"`
	TransactionRootHash string  `json:"txRootHash" gorm:"column:txroothash"`
	ReceiptRootHash     string  `json:"receiptRootHash" gorm:"column:receiptroothash"`
	ExtraData           []byte  `json:"extraData" gorm:"column:extradata"`
}

// MarshalBinary - Implementing binary marshalling function, to be invoked
// by redis before publishing data on channel
func (b *Block) MarshalBinary() ([]byte, error) {
	return json.Marshal(b)
}

// MarshalJSON - Custom JSON encoder
func (b *Block) MarshalJSON() ([]byte, error) {

	extraData := ""
	if _h := hex.EncodeToString(b.ExtraData); _h != "" {
		extraData = fmt.Sprintf("0x%s", _h)
	}

	return []byte(fmt.Sprintf(`{"hash":%q,"number":%d,"time":%d,"parentHash":%q,"difficulty":%q,"gasUsed":%d,"gasLimit":%d,"nonce":%q,"miner":%q,"size":%f,"stateRootHash":%q,"uncleHash":%q,"txRootHash":%q,"receiptRootHash":%q,"extraData":%q}`,
		b.Hash,
		b.Number,
		b.Time,
		b.ParentHash,
		b.Difficulty,
		b.GasUsed,
		b.GasLimit,
		b.Nonce,
		b.Miner,
		b.Size,
		b.StateRootHash,
		b.UncleHash,
		b.TransactionRootHash,
		b.ReceiptRootHash,
		extraData)), nil

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
