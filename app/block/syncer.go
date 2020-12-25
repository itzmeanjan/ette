package block

import (
	"log"
	"runtime"
	"sync"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gammazero/workerpool"
	"github.com/go-redis/redis/v8"
	d "github.com/itzmeanjan/ette/app/data"
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

	log.Printf("[+] Historical Data Syncing Completed\n")
}
