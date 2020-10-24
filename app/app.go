package app

import (
	"log"

	"github.com/ethereum/go-ethereum/ethclient"
	cfg "github.com/itzmeanjan/ette/app/config"
	"github.com/itzmeanjan/ette/app/db"
	"gorm.io/gorm"
)

// Setting ground up
func bootstrap(file string) (*ethclient.Client, *gorm.DB) {
	err := cfg.Read(file)
	if err != nil {
		log.Fatalln("[!] ", err)
	}

	_client := getClient()
	_db := db.Connect()

	return _client, _db
}

// Run - Application to be invoked from main runner using this function
func Run(file string) {
	_client, _ := bootstrap(file)

	subscribeToNewBlocks(_client)
}
