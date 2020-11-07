package app

import (
	"log"
	"sync"

	"github.com/adjust/rmq/v3"
	"github.com/ethereum/go-ethereum/ethclient"
	blk "github.com/itzmeanjan/ette/app/block"
	cfg "github.com/itzmeanjan/ette/app/config"
	d "github.com/itzmeanjan/ette/app/data"
	"github.com/itzmeanjan/ette/app/db"
	"github.com/itzmeanjan/ette/app/rest"
	"gorm.io/gorm"
)

// Setting ground up
func bootstrap(file string) (*ethclient.Client, rmq.Connection, *gorm.DB, *sync.Mutex, *d.SyncState) {
	err := cfg.Read(file)
	if err != nil {
		log.Fatalf("[!] Failed to read `.env` : %s\n", err.Error())
	}

	_client := getClient()
	_conn := getRedisConnection()
	_db := db.Connect()

	_lock := &sync.Mutex{}
	_synced := &d.SyncState{Synced: false}

	return _client, _conn, _db, _lock, _synced
}

// Run - Application to be invoked from main runner using this function
func Run(file string) {
	_client, _conn, _db, _lock, _synced := bootstrap(file)

	_blockQueue := getRedisMessageQueue(_conn, "block")
	defer _blockQueue.Destroy()

	// Pushing block header propagation listener to another thread of execution
	go blk.SubscribeToNewBlocks(_client, _db, _lock, _synced, _blockQueue)

	// Starting http server on main thread

	rest.RunHTTPServer(_db, _lock, _synced, _blockQueue)
}
