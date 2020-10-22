package app

import (
	"log"

	"github.com/ethereum/go-ethereum/ethclient"
)

// Setting ground up
func bootstrap(file string) *ethclient.Client{
	err := read(file)
	if err != nil {
		log.Fatalln("[!] ", err)
	}

	return getClient()
}

// Run - Application to be invoked from main runner using this function
func Run(file string) {
	client := bootstrap(file)

	subscribeToNewBlocks(client)
}
