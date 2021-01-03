package block

import (
	"log"
	"runtime"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gammazero/workerpool"
	"github.com/go-redis/redis/v8"
	"github.com/gookit/color"
	d "github.com/itzmeanjan/ette/app/data"
	"github.com/itzmeanjan/ette/app/db"
	"gorm.io/gorm"
)

// Syncer - Given ascending block number range i.e. fromBlock <= toBlock
// fetches blocks in order {fromBlock, toBlock, fromBlock + 1, toBlock - 1, fromBlock + 2, toBlock - 2 ...}
// while running n workers concurrently, where n = number of cores this machine has
//
// Waits for all of them to complete
func Syncer(client *ethclient.Client, _db *gorm.DB, redisClient *redis.Client, redisKey string, fromBlock uint64, toBlock uint64, _lock *sync.Mutex, _synced *d.SyncState, jd func(*workerpool.WorkerPool, *d.Job)) {
	if !(fromBlock <= toBlock) {
		log.Print(color.Red.Sprintf("[!] Bad block range for syncer"))
		return
	}

	wp := workerpool.New(runtime.NumCPU())
	i := fromBlock
	j := toBlock

	// Jobs need to be submitted using this interface, while
	// just mentioning which block needs to be fetched
	job := func(num uint64) {
		jd(wp, &d.Job{
			Client:      client,
			DB:          _db,
			RedisClient: redisClient,
			RedisKey:    redisKey,
			Block:       num,
			Lock:        _lock,
			Synced:      _synced,
		})
	}

	for i <= j {
		// This condition to be arrived at when range has odd number of elements
		if i == j {
			job(i)
		} else {
			job(i)
			job(j)
		}

		i++
		j--
	}

	wp.StopWait()
}

// SyncBlocksByRange - Fetch & persist all blocks in range(fromBlock, toBlock), both inclusive
//
// Range can be either ascending or descending, depending upon that proper arguments to be
// passed to `Syncer` function during invokation
func SyncBlocksByRange(client *ethclient.Client, _db *gorm.DB, redisClient *redis.Client, redisKey string, fromBlock uint64, toBlock uint64, _lock *sync.Mutex, _synced *d.SyncState) {

	// Job to be submitted and executed by each worker
	//
	// Job specification is provided in `Job` struct
	job := func(wp *workerpool.WorkerPool, j *d.Job) {
		wp.Submit(func() {

			fetchBlockByNumber(j.Client, j.Block, j.DB, j.RedisClient, j.RedisKey, j.Lock, j.Synced)

		})
	}

	log.Printf("[*] Starting block syncer\n")

	if fromBlock < toBlock {
		Syncer(client, _db, redisClient, redisKey, fromBlock, toBlock, _lock, _synced, job)
	} else {
		Syncer(client, _db, redisClient, redisKey, toBlock, fromBlock, _lock, _synced, job)
	}

	log.Printf("[+] Stopping block syncer\n")

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
		log.Printf("[*] Starting missing block finder\n")

		currentBlockNumber := db.GetCurrentBlockNumber(_db)

		// Safely reading shared variable
		_lock.Lock()
		blockCount := _synced.BlockCountInDB()
		_lock.Unlock()

		// If all blocks present in between 0 to latest block in network
		// `ette` sleeps for 1 minute & again get to work
		if currentBlockNumber+1 == blockCount {
			log.Print(color.Green.Sprintf("[+] No missing blocks found"))
			sleep()
			continue
		}

		// Job to be submitted and executed by each worker
		//
		// Job specification is provided in `Job` struct
		job := func(wp *workerpool.WorkerPool, j *d.Job) {
			wp.Submit(func() {

				// Worker fetches block by number from local storage
				block := db.GetBlock(j.DB, j.Block)
				if block == nil {
					// If not found, block fetching cycle is run, for this block
					fetchBlockByNumber(j.Client, j.Block, j.DB, j.RedisClient, j.RedisKey, j.Lock, j.Synced)
				}

			})
		}

		Syncer(client, _db, redisClient, redisKey, 0, currentBlockNumber, _lock, _synced, job)

		log.Printf("[+] Stopping missing block finder\n")
		sleep()
	}

}
