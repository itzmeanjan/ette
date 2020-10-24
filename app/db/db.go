package db

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// Blocks - ...
type Blocks struct {
	Hash       common.Hash
	Number     *big.Int
	Time       uint64
	ParentHash common.Hash
	Difficulty *big.Int
	GasUsed    uint64
	GasLimit   uint64
	Nonce      uint64
}

// Transactions - ...
type Transactions struct {
	Hash      common.Hash
	From      common.Address
	To        common.Address
	Gas       uint64
	GasPrice  *big.Int
	Cost      *big.Int
	Nonce     uint64
	State     uint8
	BlockHash common.Hash
}
