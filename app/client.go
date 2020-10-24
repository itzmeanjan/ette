package app

import (
	"log"

	"github.com/ethereum/go-ethereum/ethclient"
)

// Connect to blockchain node
func getClient() *ethclient.Client {
	client, err := ethclient.Dial(Get("RPC"))
	if err != nil {
		log.Fatalln("[!] ", err)
	}

	return client
}
