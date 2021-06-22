package app

import (
	"context"
	"log"
	"sync"

	"github.com/go-redis/redis/v8"
	cfg "github.com/itzmeanjan/ette/app/config"
	d "github.com/itzmeanjan/ette/app/data"
	"github.com/itzmeanjan/ette/app/db"
	q "github.com/itzmeanjan/ette/app/queue"
	"github.com/itzmeanjan/ette/app/rest/graph"
	"gorm.io/gorm"
)

// Setting ground up i.e. acquiring resources required & determining with
// some basic checks whether we can proceed to next step or not
func bootstrap(configFile, subscriptionPlansFile string) (*d.BlockChainNodeConnection, *redis.Client, *d.RedisInfo, *gorm.DB, *d.StatusHolder, *q.BlockProcessorQueue) {

	err := cfg.Read(configFile)
	if err != nil {
		log.Fatalf("[!] Failed to read `.env` : %s\n", err.Error())
	}

	if !(cfg.Get("EtteMode") == "HISTORICAL" || cfg.Get("EtteMode") == "REALTIME" || cfg.Get("EtteMode") == "HISTORICAL_AND_REALTIME" || cfg.Get("EtteMode") == "WRITE_SNAPSHOT" || cfg.Get("EtteMode") == "READ_SNAPSHOT") {
		log.Fatalf("[!] Failed to find `EtteMode` in configuration file\n")
	}

	// Maintaining both HTTP & Websocket based connection to blockchain
	_connection := &d.BlockChainNodeConnection{
		RPC:       getClient(true),
		Websocket: getClient(false),
	}

	_redisClient := getRedisClient()

	if _redisClient == nil {
		log.Fatalf("[!] Failed to connect to Redis Server\n")
	}

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
		State: &d.SyncState{
			BlockCountAtStartUp:     db.GetBlockCount(_db),
			MaxBlockNumberAtStartUp: db.GetCurrentBlockNumber(_db),
		},
		Mutex: &sync.RWMutex{},
	}

	_redisInfo := &d.RedisInfo{
		Client:            _redisClient,
		BlockPublishTopic: "block",
		TxPublishTopic:    "transaction",
		EventPublishTopic: "event",
	}

	// This is block processor queue
	_queue := q.New(db.GetCurrentBlockNumber(_db))

	return _connection, _redisClient, _redisInfo, _db, _status, _queue

}
