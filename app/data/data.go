package data

import (
	"encoding/json"
	"log"
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
	Hash     string `json:"hash" gorm:"column:hash"`
	From     string `json:"from" gorm:"column:from"`
	To       string `json:"to" gorm:"column:to"`
	Contract string `json:"contract" gorm:"column:contract"`
	Gas      uint64 `json:"gas" gorm:"column:gas"`
	GasPrice string `json:"gasPrice" gorm:"column:gasprice"`
	Cost     string `json:"cost" gorm:"column:cost"`
	Nonce    uint64 `json:"nonce" gorm:"column:nonce"`
	State    uint64 `json:"state" gorm:"column:state"`
}

// Transactions - Multiple transactions holder struct
type Transactions struct {
	Transactions []Transaction `json:"transactions"`
}

// ToJSON - Encoding into JSON, to be invoked when delivering to client
func (t *Transactions) ToJSON() []byte {
	data, err := json.Marshal(t)
	if err != nil {
		log.Printf("[!] Failed to encode transaction data to JSON : %s\n", err.Error())
		return nil
	}

	return data
}
