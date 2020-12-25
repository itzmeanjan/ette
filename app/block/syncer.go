package block

import (
	"runtime"
	"sync"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gammazero/workerpool"
	"github.com/go-redis/redis/v8"
	d "github.com/itzmeanjan/ette/app/data"
	"github.com/itzmeanjan/ette/app/db"
	"gorm.io/gorm"
)

// SyncBlocksByRange - Fetch & persist all blocks in range(fromBlock, toBlock), both inclusive
//
// This function can be called for syncing either in forward/ backward direction, depending upon
// parameter passed in for `fromBlock` & `toBlock` field
func SyncBlocksByRange(client *ethclient.Client, _db *gorm.DB, redisClient *redis.Client, redisKey string, fromBlock uint64, toBlock uint64, _lock *sync.Mutex, _synced *d.SyncState) {
	wp := workerpool.New(runtime.NumCPU())

	if fromBlock < toBlock {
		for i := fromBlock; i <= toBlock; i++ {

			func(num uint64) {
				wp.Submit(func() {
					fetchBlockByNumber(client, num, _db, redisClient, redisKey, _lock, _synced)
				})
			}(i)

		}
	} else {
		for i := fromBlock; i >= toBlock; i-- {

			func(num uint64) {
				wp.Submit(func() {
					fetchBlockByNumber(client, num, _db, redisClient, redisKey, _lock, _synced)
				})
			}(i)

		}
	}

	wp.StopWait()

	// If traversing in backward direction, then try checking lowest block number present in DB
	// and try to reprocess upto 0, if not reached 0 yet
	if fromBlock > toBlock {

		lowestBlockNumber := db.GetCurrentOldestBlockNumber(_db)
		if lowestBlockNumber != 0 {
			go SyncBlocksByRange(client, _db, redisClient, redisKey, lowestBlockNumber, 0, _lock, _synced)
		} else {

			currentBlockNumber := db.GetCurrentBlockNumber(_db)
			blockCount := db.GetBlockCount(_db)
			if currentBlockNumber-lowestBlockNumber+1 == blockCount {
				return
			}

		}

	}
}
