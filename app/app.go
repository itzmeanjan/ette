package app

import (
	"context"
	"log"
	"sync"

	"github.com/go-redis/redis/v8"
	blk "github.com/itzmeanjan/ette/app/block"
	cfg "github.com/itzmeanjan/ette/app/config"
	d "github.com/itzmeanjan/ette/app/data"
	"github.com/itzmeanjan/ette/app/db"
	"github.com/itzmeanjan/ette/app/rest"
	"github.com/itzmeanjan/ette/app/rest/graph"
	"gorm.io/gorm"
)

// Setting ground up
func bootstrap(configFile, subscriptionPlansFile string) (*d.BlockChainNodeConnection, *redis.Client, *gorm.DB, *d.StatusHolder) {
	err := cfg.Read(configFile)
	if err != nil {
		log.Fatalf("[!] Failed to read `.env` : %s\n", err.Error())
	}

	if !(cfg.Get("EtteMode") == "1" || cfg.Get("EtteMode") == "2" || cfg.Get("EtteMode") == "3") {
		log.Fatalf("[!] Failed to find `EtteMode` in configuration file\n")
	}

	// Maintaining both HTTP & Websocket based connection to blockchain
	_connection := &d.BlockChainNodeConnection{
		RPC:       getClient(true),
		Websocket: getClient(false),
	}

	_redisClient := getPubSubClient()

	if err := _redisClient.FlushAll(context.Background()).Err(); err != nil {
		log.Printf("[!] Failed to flush all keys from redis : %s\n", err.Error())
	}

	_db := db.Connect()

	// Populating subscription plans from `.plans.json` into
	// database table, at application start up
	db.PersistAllSubscriptionPlans(_db, subscriptionPlansFile)

	// Passing db handle, to graph package, so that it can be used
	// for resolving graphQL queries
	graph.GetDatabaseConnection(_db)

	_status := &d.StatusHolder{
		State: &d.SyncState{BlockCountAtStartUp: db.GetBlockCount(_db)},
		Mutex: &sync.RWMutex{},
	}

	return _connection, _redisClient, _db, _status
}

// Run - Application to be invoked from main runner using this function
func Run(configFile, subscriptionPlansFile string) {
	_connection, _redisClient, _db, _status := bootstrap(configFile, subscriptionPlansFile)
	_redisInfo := d.RedisInfo{
		Client:    _redisClient,
		QueueName: "blocks",
	}

	// Pushing block header propagation listener to another thread of execution
	go blk.SubscribeToNewBlocks(_connection, _db, _status, &_redisInfo)

	// Starting http server on main thread
	rest.RunHTTPServer(_db, _status, _redisClient)
}
