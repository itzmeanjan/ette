package block

import (
	"context"
	"log"
	"math/big"
	"runtime"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gammazero/workerpool"
	"github.com/itzmeanjan/ette/app/db"
	"gorm.io/gorm"
)

// SyncToLatestBlock - Fetch all blocks upto latest block
func SyncToLatestBlock(client *ethclient.Client, _db *gorm.DB) {
	latestBlockNum, err := client.BlockNumber(context.Background())
	if err != nil {
		log.Fatalln("[!] ", err)
	}

	wp := workerpool.New(runtime.NumCPU())

	for i := uint64(0); i < latestBlockNum; i++ {
		blockNum := i

		wp.Submit(func() {
			fetchBlockByNumber(client, blockNum, _db)
		})
	}

	wp.StopWait()
}

// SubscribeToNewBlocks - Listen for event when new block header is
// available, then fetch block content ( including all transactions )
// in different worker
func SubscribeToNewBlocks(client *ethclient.Client, _db *gorm.DB) {
	headerChan := make(chan *types.Header)

	subs, err := client.SubscribeNewHead(context.Background(), headerChan)
	if err != nil {
		log.Fatalln("[!] ", err)
	}

	// Scheduling unsubscribe, to be executed when end of this block is reached
	defer subs.Unsubscribe()

	for {
		select {
		case err := <-subs.Err():
			log.Println("[!] ", err)
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
		log.Println("[!] ", err)
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
		log.Println("[!] ", err)
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
		log.Println("[!] Empty Block : ", block.NumberU64())
		return
	}

	for _, v := range block.Transactions() {
		fetchTransactionByHash(client, block, v, _db)
	}
}

// Fetching specific transaction related data & persisting in database
func fetchTransactionByHash(client *ethclient.Client, block *types.Block, tx *types.Transaction, _db *gorm.DB) {
	// if DB entry already exists for this tx
	if res := db.GetTransaction(_db, block.Hash(), tx.Hash()); res != nil {
		return
	}

	receipt, err := client.TransactionReceipt(context.Background(), tx.Hash())
	if err != nil {
		log.Println("[!] ", err)
		return
	}

	sender, err := client.TransactionSender(context.Background(), tx, block.Hash(), receipt.TransactionIndex)
	if err != nil {
		log.Println("[!] ", err)
		return
	}

	db.PutTransaction(_db, tx, receipt, sender)
	db.PutEvent(_db, receipt)
	log.Println(sender.Hex(), tx.To().Hex(), "[ ", block.Number().String(), " ]")
}
