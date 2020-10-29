package app

import (
	"log"
	"sync"

	"github.com/ethereum/go-ethereum/ethclient"
	blk "github.com/itzmeanjan/ette/app/block"
	d "github.com/itzmeanjan/ette/app/data"
	cfg "github.com/itzmeanjan/ette/app/config"
	"github.com/itzmeanjan/ette/app/db"
	"gorm.io/gorm"
)

// Setting ground up
func bootstrap(file string) (*ethclient.Client, *gorm.DB, *sync.Mutex, *d.SyncState) {
	err := cfg.Read(file)
	if err != nil {
		log.Fatalf("[!] Failed to read `.env` : %s\n", err.Error())
	}

	_client := getClient()
	_db := db.Connect()

	_lock := &sync.Mutex{}
	_synced := &d.SyncState{Synced: false}

	return _client, _db, _lock, _synced
}

// Run - Application to be invoked from main runner using this function
func Run(file string) {
	_client, _db, _lock, _synced := bootstrap(file)

	go blk.SyncToLatestBlock(_client, _db, _lock, _synced)
	blk.SubscribeToNewBlocks(_client, _db)
}
