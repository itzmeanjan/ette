package db

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// Blocks - ...
type Blocks struct {
	Hash       string `gorm:"column:hash;type:char(66);primaryKey"`
	Number     string `gorm:"column:number;type:varchar;not null"`
	Time       uint64 `gorm:"column:time;type:bigint;not null"`
	ParentHash string `gorm:"column:parenthash;type:char(66);not null"`
	Difficulty string `gorm:"column:difficulty;type:varchar;not null"`
	GasUsed    uint64 `gorm:"column:gasused;type:bigint;not null"`
	GasLimit   uint64 `gorm:"column:gaslimit;type:bigint;not null"`
	Nonce      uint64 `gorm:"column:nonce;type:bigint;not null"`
}

// Transactions - ...
type Transactions struct {
	Hash      common.Hash    `gorm:"column:hash;type:char(66);primaryKey"`
	From      common.Address `gorm:"column:from;type:char(42);not null"`
	To        common.Address `gorm:"column:to;type:char(42); not null"`
	Gas       uint64         `gorm:"column:gas;type:bigint;not null"`
	GasPrice  *big.Int       `gorm:"column:gasprice;type:varchar;not null"`
	Cost      *big.Int       `gorm:"column:cost;type:varchar;not null"`
	Nonce     uint64         `gorm:"column:nonce;type:bigint;not null"`
	State     uint8          `gorm:"column:state;type:smallint;not null"`
	BlockHash common.Hash    `gorm:"column:blockhash;type:char(66);not null"`
}
