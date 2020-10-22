package app

import (
	"context"
	"log"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

func subscribeToNewBlocks(client *ethclient.Client) {
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

	if block.Transactions().Len() == 0 {
		log.Println("[!] Empty Block : ", block.NumberU64())
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
