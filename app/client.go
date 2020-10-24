package app

import (
	"log"

	"github.com/ethereum/go-ethereum/ethclient"
	cfg "github.com/itzmeanjan/ette/app/config"
)

// Connect to blockchain node
func getClient() *ethclient.Client {
	client, err := ethclient.Dial(cfg.Get("RPC"))
	if err != nil {
		log.Fatalln("[!] ", err)
	}

	return client
}
