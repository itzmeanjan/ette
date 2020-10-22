package app

import (
	"context"
	"log"

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
			log.Println(header.Hash().Hex())
		}
	}
}
