package app

import (
	"log"

	"github.com/ethereum/go-ethereum/ethclient"
)

// Connect to blockchain node
func getClient() *ethclient.Client{
	client, err := ethclient.Dial(get("RPC"))
	if err != nil {
		log.Fatalln("[!] ", err)
		return nil
	}

	return client
}
