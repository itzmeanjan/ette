package block

import (
	"context"
	"log"
	"math/big"
	"os"
	"runtime"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gammazero/workerpool"
)

// SyncToLatestBlock - Fetch all blocks upto latest block
func SyncToLatestBlock(client *ethclient.Client) {
	latestBlockNum, err := client.BlockNumber(context.Background())
	if err != nil {
		log.Fatalln("[!] ", err)
	}

	wp := workerpool.New(runtime.NumCPU())

	for i := uint64(0); i < latestBlockNum; i++ {
		blockNum := i

		wp.Submit(func() {
			fetchBlockByNumber(client, blockNum)
		})
	}

	wp.StopWait()
}

// SubscribeToNewBlocks - Listen for event when new block header is
// available, then fetch block content ( including all transactions )
// in different worker
func SubscribeToNewBlocks(client *ethclient.Client) {
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
			go fetchBlockByHash(client, header.Hash())
		}
	}
}

// Fetching block content using blockHash
func fetchBlockByHash(client *ethclient.Client, hash common.Hash) {
	block, err := client.BlockByHash(context.Background(), hash)
	if err != nil {
		log.Println("[!] ", err)
		return
	}

	fetchBlockContent(client, block)
}

// Fetching block content using block number
func fetchBlockByNumber(client *ethclient.Client, number uint64) {
	_num := big.NewInt(0)
	_num = _num.SetUint64(number)

	block, err := client.BlockByNumber(context.Background(), _num)
	if err != nil {
		log.Println("[!] ", err)
		return
	}

	fetchBlockContent(client, block)
}

// Fetching all transactions in this block, along with their receipt
func fetchBlockContent(client *ethclient.Client, block *types.Block) {
	if block.Transactions().Len() == 0 {
		log.Println("[!] Empty Block : ", block.NumberU64())
		return
	}

	for _, v := range block.Transactions() {
		receipt, err := client.TransactionReceipt(context.Background(), v.Hash())
		if err != nil {
			log.Println("[!] ", err)
			continue
		}

		sender, err := client.TransactionSender(context.Background(), v, block.Hash(), receipt.TransactionIndex)
		if err != nil {
			log.Println("[!] ", err)
			continue
		}

		log.Println(sender.Hex(), v.To().Hex(), "[ ", block.NumberU64(), " ]")
	}
}
