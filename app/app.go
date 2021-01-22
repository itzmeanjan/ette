package app

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gookit/color"
	blk "github.com/itzmeanjan/ette/app/block"
	cfg "github.com/itzmeanjan/ette/app/config"
	d "github.com/itzmeanjan/ette/app/data"
	"github.com/itzmeanjan/ette/app/db"
	"github.com/itzmeanjan/ette/app/rest"
	"github.com/itzmeanjan/ette/app/rest/graph"
	ss "github.com/itzmeanjan/ette/app/snapshot"
	"gorm.io/gorm"
)

// Setting ground up
func bootstrap(configFile, subscriptionPlansFile string) (*d.BlockChainNodeConnection, *redis.Client, *gorm.DB, *d.StatusHolder) {
	err := cfg.Read(configFile)
	if err != nil {
		log.Fatalf("[!] Failed to read `.env` : %s\n", err.Error())
	}

	if !(cfg.Get("EtteMode") == "1" || cfg.Get("EtteMode") == "2" || cfg.Get("EtteMode") == "3" || cfg.Get("EtteMode") == "4") {
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
		State: &d.SyncState{BlockCountAtStartUp: db.GetBlockCount(_db)},
		Mutex: &sync.RWMutex{},
	}

	return _connection, _redisClient, _db, _status
}

// Run - Application to be invoked from main runner using this function
func Run(configFile, subscriptionPlansFile string) {
	_connection, _redisClient, _db, _status := bootstrap(configFile, subscriptionPlansFile)
	_redisInfo := d.RedisInfo{
		Client:                     _redisClient,
		BlockRetryQueueName:        "blocks_in_retry_queue",
		UnfinalizedBlocksQueueName: "unfinalized_blocks",
	}

	// Attempting to listen to Ctrl+C signal
	// and when received gracefully shutting down `ette`
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)

	// All resources being used gets cleaned up
	// when we're returning from this function scope
	go func() {

		<-interruptChan

		sql, err := _db.DB()
		if err != nil {
			log.Printf(color.Red.Sprintf("[!] Failed to get underlying DB connection : %s", err.Error()))
			return
		}

		if err := sql.Close(); err != nil {
			log.Printf(color.Red.Sprintf("[!] Failed to close underlying DB connection : %s", err.Error()))
			return
		}

		if err := _redisInfo.Client.Close(); err != nil {
			log.Printf(color.Red.Sprintf("[!] Failed to close connection to Redis : %s", err.Error()))
			return
		}

		// Stopping process
		log.Printf(color.Magenta.Sprintf("\n[+] Gracefully shut down `ette`"))
		os.Exit(0)

	}()

	// User has requested `ette` to take a snapshot of current database state
	if cfg.Get("EtteMode") == "4" {

		// checking if there's anything to snapshot or not
		if _status.BlockCountInDB() == 0 {
			log.Printf("[*] Nothing to snapshot\n")
			return
		}

		// this is the file snapshot to be taken
		_snapshotFile := cfg.GetSnapshotFile()
		_start := time.Now().UTC()

		log.Printf("[*] Starting snapshotting at : %s [ Sink : %s ]\n", _start, _snapshotFile)

		// taking snapshot, this might take some time
		_ret := ss.TakeSnapshot(_db, _snapshotFile, db.GetCurrentOldestBlockNumber(_db), db.GetCurrentBlockNumber(_db), _status.BlockCountInDB())
		if _ret {
			log.Printf("[+] Snapshotted in : %s\n", time.Now().UTC().Sub(_start))
		} else {
			log.Printf("[!] Snapshotting failed in : %s\n", time.Now().UTC().Sub(_start))
		}

		return

	}

	// Pushing block header propagation listener to another thread of execution
	go blk.SubscribeToNewBlocks(_connection, _db, _status, &_redisInfo)

	// Starting http server on main thread
	rest.RunHTTPServer(_db, _status, _redisClient)
}
