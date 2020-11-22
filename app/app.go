package app

import (
	"context"
	"log"
	"sync"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-redis/redis/v8"
	blk "github.com/itzmeanjan/ette/app/block"
	cfg "github.com/itzmeanjan/ette/app/config"
	d "github.com/itzmeanjan/ette/app/data"
	"github.com/itzmeanjan/ette/app/db"
	"github.com/itzmeanjan/ette/app/rest"
	"gorm.io/gorm"
)

// Setting ground up
func bootstrap(file string) (*ethclient.Client, *redis.Client, *gorm.DB, *sync.Mutex, *d.SyncState) {
	err := cfg.Read(file)
	if err != nil {
		log.Fatalf("[!] Failed to read `.env` : %s\n", err.Error())
	}

	_client := getClient()
	_redisClient := getPubSubClient()

	if err := _redisClient.FlushAll(context.Background()).Err(); err != nil {
		log.Printf("[!] Failed to flush all keys from redis : %s\n", err.Error())
	}

	_db := db.Connect()

	_lock := &sync.Mutex{}
	_synced := &d.SyncState{Target: 0, Done: 0}

	return _client, _redisClient, _db, _lock, _synced
}

// Run - Application to be invoked from main runner using this function
func Run(file string) {
	_client, _redisClient, _db, _lock, _synced := bootstrap(file)

	// Pushing block header propagation listener to another thread of execution
	go blk.SubscribeToNewBlocks(_client, _db, _lock, _synced, _redisClient)

	// Starting http server on main thread
	rest.RunHTTPServer(_db, _lock, _synced, _redisClient)
}
