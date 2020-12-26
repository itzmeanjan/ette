package block

import (
	"log"
	"runtime"
	"sync"
	"time"

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
	log.Printf("[*] Starting backward syncing\n")
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
		i := fromBlock
		j := toBlock

		// Trying to reach middle of range (fromBlock, toBlock)
		//
		// If starting block is 100 & ending is 1, in each iteration
		// two workers to be started i.e. for fetching 100, 1 block numbers
		//
		// In next iteration, it'll start worker for fetching 99 & 2; 98 & 3 ...
		// Will stop when reaches mid of range i.e. 50, 51
		for i >= j {

			// This condition to be arrived at when range has odd number of elements
			//
			// This will be very last state, i.e. when it's time to get out of loop
			if i == j {
				func(num uint64) {
					wp.Submit(func() {
						fetchBlockByNumber(client, num, _db, redisClient, redisKey, _lock, _synced)
					})
				}(i)
			} else {
				func(num uint64) {
					wp.Submit(func() {
						fetchBlockByNumber(client, num, _db, redisClient, redisKey, _lock, _synced)
					})
				}(i)

				func(num uint64) {
					wp.Submit(func() {
						fetchBlockByNumber(client, num, _db, redisClient, redisKey, _lock, _synced)
					})
				}(j)
			}

			i--
			j++

		}
	}

	wp.StopWait()
	log.Printf("[+] Completed backward syncing\n")

	// Once completed first iteration of processing blocks upto last time where it left
	// off, we're going to start worker to look at DB & decide which blocks are missing
	// i.e. need to be fetched again
	//
	// And this will itself run as a infinite job, completes one iteration &
	// takes break for 1 min, then repeats
	go SyncMissingBlocksInDB(client, _db, redisClient, redisKey, _lock, _synced)
}

// SyncMissingBlocksInDB - Checks with database for what blocks are present & what are not, fetches missing
// blocks & related data iteratively
func SyncMissingBlocksInDB(client *ethclient.Client, _db *gorm.DB, redisClient *redis.Client, redisKey string, _lock *sync.Mutex, _synced *d.SyncState) {

	// Sleep for 1 minute & then again check whether we need to fetch missing blocks or not
	sleep := func() {
		time.Sleep(time.Duration(1) * time.Minute)
	}

	for {
		log.Printf("[*] Starting missing block finder")
		currentBlockNumber := db.GetCurrentBlockNumber(_db)

		_lock.Lock()
		blockCount := _synced.StartedWith + _synced.Done
		_lock.Unlock()

		// If all blocks present in between 0 to latest block in network
		// `ette` sleeps for 1 minute & again get to work
		if currentBlockNumber+1 == blockCount {
			sleep()
			continue
		}

		// Starting worker pool to leverage multi core architecture of machine
		wp := workerpool.New(runtime.NumCPU())

		var i uint64
		j := currentBlockNumber

		// Trying to process range faster by processing
		// from both sides of range, one pointer moving from
		// end of range; another one moving from start of range
		// while both of them trying to reach center of range
		//
		// And iteration stops as soon as we reach center of range
		for i <= j {

			if i == j {

				func(num uint64) {

					block := db.GetBlockByNumber(_db, num)
					if block == nil {
						wp.Submit(func() {
							fetchBlockByNumber(client, num, _db, redisClient, redisKey, _lock, _synced)
						})
					}

				}(i)

			} else {
				func(num uint64) {

					block := db.GetBlockByNumber(_db, num)
					if block == nil {
						wp.Submit(func() {
							fetchBlockByNumber(client, num, _db, redisClient, redisKey, _lock, _synced)
						})
					}

				}(i)

				func(num uint64) {

					block := db.GetBlockByNumber(_db, num)
					if block == nil {
						wp.Submit(func() {
							fetchBlockByNumber(client, num, _db, redisClient, redisKey, _lock, _synced)
						})
					}

				}(j)
			}

			i++
			j--

		}

		wp.StopWait()
		log.Printf("[+] Stopping missing block finder")

		sleep()
	}

}
