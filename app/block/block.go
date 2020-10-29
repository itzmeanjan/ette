package block

import (
	"context"
	"log"
	"math/big"
	"runtime"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gammazero/workerpool"
	"github.com/itzmeanjan/ette/app"
	"github.com/itzmeanjan/ette/app/db"
	"gorm.io/gorm"
)

// SyncToLatestBlock - Fetch all blocks upto latest block
func SyncToLatestBlock(client *ethclient.Client, _db *gorm.DB, _lock *sync.Mutex, _synced *app.SyncState) {
	latestBlockNum, err := client.BlockNumber(context.Background())
	if err != nil {
		log.Fatalf("[!] Failed to fetch latest  block number : %s\n", err.Error())
	}

	wp := workerpool.New(runtime.NumCPU())

	for i := uint64(0); i < latestBlockNum; i++ {

		func(blockNum uint64) {
			wp.Submit(func() {
				fetchBlockByNumber(client, blockNum, _db)
			})
		}(i)

	}

	wp.StopWait()

	// Safely updating sync state
	_lock.Lock()
	_synced.Synced = true
	_lock.Unlock()

	log.Printf("[+] Syncing Completed\n")
}

// SubscribeToNewBlocks - Listen for event when new block header is
// available, then fetch block content ( including all transactions )
// in different worker
func SubscribeToNewBlocks(client *ethclient.Client, _db *gorm.DB) {
	headerChan := make(chan *types.Header)

	subs, err := client.SubscribeNewHead(context.Background(), headerChan)
	if err != nil {
		log.Fatalf("[!] Failed to subscribe to block headers : %s\n", err.Error())
	}

	// Scheduling unsubscribe, to be executed when end of this block is reached
	defer subs.Unsubscribe()

	for {
		select {
		case err := <-subs.Err():
			log.Printf("[!] Block header subscription failed in mid : %s\n", err.Error())
			break
		case header := <-headerChan:
			go fetchBlockByHash(client, header.Hash(), _db)
		}
	}
}

// Fetching block content using blockHash
func fetchBlockByHash(client *ethclient.Client, hash common.Hash, _db *gorm.DB) {
	block, err := client.BlockByHash(context.Background(), hash)
	if err != nil {
		log.Printf("[!] Failed to fetch block by hash : %s\n", err.Error())
		return
	}

	db.PutBlock(_db, block)
	fetchBlockContent(client, block, _db)
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

	fetchBlockContent(client, block, _db)
}

// Fetching all transactions in this block, along with their receipt
func fetchBlockContent(client *ethclient.Client, block *types.Block, _db *gorm.DB) {
	if block.Transactions().Len() == 0 {
		log.Printf("[!] Empty Block : %d\n", block.NumberU64())
		return
	}

	for _, v := range block.Transactions() {
		fetchTransactionByHash(client, block, v, _db)
	}
}

// Fetching specific transaction related data & persisting in database
func fetchTransactionByHash(client *ethclient.Client, block *types.Block, tx *types.Transaction, _db *gorm.DB) {
	// If DB entry already exists for this tx
	if res := db.GetTransaction(_db, block.Hash(), tx.Hash()); res != nil {
		return
	}

	receipt, err := client.TransactionReceipt(context.Background(), tx.Hash())
	if err != nil {
		log.Printf("[!] Failed to fetch tx receipt : %s\n", err.Error())
		return
	}

	sender, err := client.TransactionSender(context.Background(), tx, block.Hash(), receipt.TransactionIndex)
	if err != nil {
		log.Printf("[!] Failed to fetch tx sender : %s\n", err.Error())
		return
	}

	db.PutTransaction(_db, tx, receipt, sender)
	db.PutEvent(_db, receipt)
	log.Printf("[+] Block %d with %d tx(s)\n", block.NumberU64(), len(block.Transactions()))
}