package block

import (
	"context"
	"log"
	"runtime"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/gammazero/workerpool"
	"github.com/gookit/color"
	cfg "github.com/itzmeanjan/ette/app/config"
	d "github.com/itzmeanjan/ette/app/data"
	"github.com/itzmeanjan/ette/app/db"
	"gorm.io/gorm"
)

// SubscribeToNewBlocks - Listen for event when new block header is
// available, then fetch block content ( including all transactions )
// in different worker
func SubscribeToNewBlocks(connection *d.BlockChainNodeConnection, _db *gorm.DB, status *d.StatusHolder, redis *d.RedisInfo) {
	headerChan := make(chan *types.Header)

	subs, err := connection.Websocket.SubscribeNewHead(context.Background(), headerChan)
	if err != nil {
		log.Fatal(color.Red.Sprintf("[!] Failed to subscribe to block headers : %s", err.Error()))
	}
	// Scheduling unsubscribe, to be executed when end of this execution scope is reached
	defer subs.Unsubscribe()

	// Last time `ette` stopped syncing here
	currentHighestBlockNumber := db.GetCurrentBlockNumber(_db)

	// Flag to check for whether this is first time block header being received or not
	//
	// If yes, we'll start syncer to fetch all block in range (last block processed, latest block)
	first := true
	// Creating a job queue of size `#-of CPUs present in machine`
	// where block fetching requests to be submitted
	wp := workerpool.New(runtime.NumCPU() * int(cfg.GetConcurrencyFactor()))
	// Scheduling worker pool closing, to be called,
	// when returning from this execution scope i.e. function
	defer wp.Stop()

	for {
		select {
		case err := <-subs.Err():
			log.Fatal(color.Red.Sprintf("[!] Listener stopped : %s", err.Error()))
			break
		case header := <-headerChan:
			if first {

				// Starting now, to be used for calculating system performance, uptime etc.
				status.SetStartedAt()

				// If historical data query features are enabled
				// only then we need to sync to latest state of block chain
				if cfg.Get("EtteMode") == "1" || cfg.Get("EtteMode") == "3" {
					// Starting syncer in another thread, where it'll keep fetching
					// blocks from highest block number it fetched last time to current network block number
					// i.e. trying to fill up gap, which was caused when `ette` was offline
					//
					// But in reverse direction i.e. from 100 to 50, where `ette` fetched upto 50 last time & 100
					// is latest block, got mined in network
					//
					// Yes it's going refetch 50, due to the fact, some portions of 50 might be missed in last try
					// So, it'll check & decide whether persisting again is required or not
					//
					// This backward traversal mechanism gives us more recent blockchain happenings to cover
					go SyncBlocksByRange(connection.RPC, _db, redis, header.Number.Uint64()-1, currentHighestBlockNumber, status)

					// Starting go routine for fetching blocks `ette` failed to process in previous attempt
					//
					// Uses Redis backed queue for fetching pending block hash & retries
					go retryBlockFetching(connection.RPC, _db, redis, status)

					// Making sure on when next latest block header is received, it'll not
					// start another syncer
				}
				first = false

			}

			// As soon as new block is mined, `ette` will try to fetch it
			// and that job will be submitted in job queue
			//
			// Putting it in a different function scope for safety purpose
			// so that job submitter gets its own copy of block number & block hash,
			// otherwise it might get wrong info, if new block gets mined very soon &
			// this job is not yet submitted
			//
			// Though it'll be picked up sometime in future ( by missing block finder ), but it can be safely handled now
			// so that it gets processed immediately
			func(blockHash common.Hash, blockNumber string) {

				wp.Submit(func() {

					FetchBlockByHash(connection.RPC,
						blockHash,
						blockNumber,
						_db,
						redis,
						status)

				})

			}(header.Hash(), header.Number.String())

		}
	}
}
