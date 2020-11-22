package block

import (
	"context"
	"log"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-redis/redis/v8"
	d "github.com/itzmeanjan/ette/app/data"
	"github.com/itzmeanjan/ette/app/db"
	"gorm.io/gorm"
)

// Fetching block content using blockHash
func fetchBlockByHash(client *ethclient.Client, hash common.Hash, _db *gorm.DB, redisClient *redis.Client, _lock *sync.Mutex, _synced *d.SyncState) {
	block, err := client.BlockByHash(context.Background(), hash)
	if err != nil {
		log.Printf("[!] Failed to fetch block by hash : %s\n", err.Error())
		return
	}

	db.PutBlock(_db, block)

	if err := redisClient.Publish(context.Background(), "block", &d.Block{
		Hash:                block.Hash().Hex(),
		Number:              block.NumberU64(),
		Time:                block.Time(),
		ParentHash:          block.ParentHash().Hex(),
		Difficulty:          block.Difficulty().String(),
		GasUsed:             block.GasUsed(),
		GasLimit:            block.GasLimit(),
		Nonce:               block.Nonce(),
		Miner:               block.Coinbase().Hex(),
		Size:                float64(block.Size()),
		TransactionRootHash: block.TxHash().Hex(),
		ReceiptRootHash:     block.ReceiptHash().Hex(),
	}).Err(); err != nil {

		log.Printf("[!] Failed to publish block %d in channel : %s\n", block.NumberU64(), err.Error())

	}

	fetchBlockContent(client, block, _db, redisClient, _lock, _synced)
}

// Fetching block content using block number
func fetchBlockByNumber(client *ethclient.Client, number uint64, _db *gorm.DB, _lock *sync.Mutex, _synced *d.SyncState) {
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

	fetchBlockContent(client, block, _db, nil, _lock, _synced)
}

// Fetching all transactions in this block, along with their receipt
func fetchBlockContent(client *ethclient.Client, block *types.Block, _db *gorm.DB, redisClient *redis.Client, _lock *sync.Mutex, _synced *d.SyncState) {
	if block.Transactions().Len() == 0 {
		log.Printf("[!] Empty Block : %d\n", block.NumberU64())

		// -- Safely updating sync state holder
		_lock.Lock()
		defer _lock.Unlock()

		_synced.Done++
		if block.NumberU64() >= _synced.Target {
			_synced.Target = block.NumberU64() + 1
		}
		// ---

		return
	}

	for _, v := range block.Transactions() {
		fetchTransactionByHash(client, block, v, _db, redisClient, _lock, _synced)
	}
}
