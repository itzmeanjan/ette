package block

import (
	"context"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-redis/redis/v8"
	"github.com/itzmeanjan/ette/app/db"
	"gorm.io/gorm"
)

// Fetching block content using blockHash
func fetchBlockByHash(client *ethclient.Client, hash common.Hash, _db *gorm.DB, redisClient *redis.Client) {
	block, err := client.BlockByHash(context.Background(), hash)
	if err != nil {
		log.Printf("[!] Failed to fetch block by hash : %s\n", err.Error())
		return
	}

	db.PutBlock(_db, block)
	fetchBlockContent(client, block, _db, redisClient)
}

// Fetching block content using block number
func fetchBlockByNumber(client *ethclient.Client, number uint64, _db *gorm.DB) {
	_num := big.NewInt(0)
	_num = _num.SetUint64(number)

	block, err := client.BlockByNumber(context.Background(), _num)
	if err != nil {
		log.Printf("[!] Failed to fetch block by number : %s\n", err)
		return
	}

	if res := db.GetBlock(_db, number); res == nil {
		db.PutBlock(_db, block)
	}

	fetchBlockContent(client, block, _db, nil)
}

// Fetching all transactions in this block, along with their receipt
func fetchBlockContent(client *ethclient.Client, block *types.Block, _db *gorm.DB, redisClient *redis.Client) {
	if block.Transactions().Len() == 0 {
		log.Printf("[!] Empty Block : %d\n", block.NumberU64())
		return
	}

	for _, v := range block.Transactions() {
		fetchTransactionByHash(client, block, v, _db, redisClient)
	}
}
