package block

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-redis/redis/v8"
	cfg "github.com/itzmeanjan/ette/app/config"
	d "github.com/itzmeanjan/ette/app/data"
	"github.com/itzmeanjan/ette/app/db"
	"gorm.io/gorm"
)

// Running a different executor for keeping block count in memory, as fresh as we can keep
func keepBlockCountInMemory(_db *gorm.DB, _lock *sync.Mutex, _synced *d.SyncState) {

	for {
		count := db.GetBlockCount(_db)
		if count == 0 {
			continue
		}

		_lock.Lock()
		_synced.BlockCountInDB = count
		_lock.Unlock()
	}

}

// Fetching block content using blockHash
func fetchBlockByHash(client *ethclient.Client, hash common.Hash, number string, _db *gorm.DB, redisClient *redis.Client, redisKey string, _lock *sync.Mutex, _synced *d.SyncState) {
	block, err := client.BlockByHash(context.Background(), hash)
	if err != nil {
		// Pushing block number into Redis queue for retrying later
		pushBlockHashIntoRedisQueue(redisClient, redisKey, number)

		log.Printf("[!] Failed to fetch block by hash [ block : %s] : %s\n", number, err.Error())
		return
	}

	// Publishes block data to all listening parties
	// on `block` channel
	publishBlock := func() {
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
	}

	// Controlling behaviour of ette depending upon value of `EtteMode`
	switch cfg.Get("EtteMode") {
	case "1":
		db.StoreBlock(_db, block)
	case "2":
		publishBlock()
	case "3":
		db.StoreBlock(_db, block)
		publishBlock()
	}

	fetchBlockContent(client, block, _db, redisClient, redisKey, true, _lock, _synced)
}

// Fetching block content using block number
func fetchBlockByNumber(client *ethclient.Client, number uint64, _db *gorm.DB, redisClient *redis.Client, redisKey string, _lock *sync.Mutex, _synced *d.SyncState) {
	_num := big.NewInt(0)
	_num = _num.SetUint64(number)

	block, err := client.BlockByNumber(context.Background(), _num)
	if err != nil {
		// Pushing block number into Redis queue for retrying later
		pushBlockHashIntoRedisQueue(redisClient, redisKey, fmt.Sprintf("%d", number))

		log.Printf("[!] Failed to fetch block by number [ block : %d ] : %s\n", number, err)
		return
	}

	// Either creates new entry or updates existing one
	if !db.StoreBlock(_db, block) {
		// Pushing block number into Redis queue for retrying later
		pushBlockHashIntoRedisQueue(redisClient, redisKey, fmt.Sprintf("%d", number))
		return
	}

	fetchBlockContent(client, block, _db, redisClient, redisKey, false, _lock, _synced)
}

// Fetching all transactions in this block, along with their receipt
func fetchBlockContent(client *ethclient.Client, block *types.Block, _db *gorm.DB, redisClient *redis.Client, redisKey string, publishable bool, _lock *sync.Mutex, _synced *d.SyncState) {
	if block.Transactions().Len() == 0 {
		log.Printf("[!] Empty Block : %d\n", block.NumberU64())
		safeUpdationOfSyncState(_lock, _synced)
		return
	}

	count := 0
	for _, v := range block.Transactions() {
		if fetchTransactionByHash(client, block, v, _db, redisClient, redisKey, publishable, _lock, _synced) {
			count++
		}
	}

	if count == len(block.Transactions()) {
		log.Printf("[+] Block %d with %d tx(s)\n", block.NumberU64(), len(block.Transactions()))
		safeUpdationOfSyncState(_lock, _synced)
		return
	}

	// Pushing block number into Redis queue for retrying later
	pushBlockHashIntoRedisQueue(redisClient, redisKey, block.Number().String())
}

// Updating shared varible between worker go routines, denoting progress of
// `ette`, in terms of data syncing
func safeUpdationOfSyncState(_lock *sync.Mutex, _synced *d.SyncState) {
	_lock.Lock()
	defer _lock.Unlock()

	_synced.Done++
}
