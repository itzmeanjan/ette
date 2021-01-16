package block

import (
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gammazero/workerpool"
	"github.com/gookit/color"
	cfg "github.com/itzmeanjan/ette/app/config"
	"github.com/itzmeanjan/ette/app/data"
	d "github.com/itzmeanjan/ette/app/data"
	"github.com/itzmeanjan/ette/app/db"
	"gorm.io/gorm"
)

// Syncer - Given ascending block number range i.e. fromBlock <= toBlock
// fetches blocks in order {fromBlock, toBlock, fromBlock + 1, toBlock - 1, fromBlock + 2, toBlock - 2 ...}
// while running n workers concurrently, where n = number of cores this machine has
//
// Waits for all of them to complete
func Syncer(client *ethclient.Client, _db *gorm.DB, redis *data.RedisInfo, fromBlock uint64, toBlock uint64, status *d.StatusHolder, jd func(*workerpool.WorkerPool, *d.Job)) {
	if !(fromBlock <= toBlock) {
		log.Print(color.Red.Sprintf("[!] Bad block range for syncer"))
		return
	}

	wp := workerpool.New(runtime.NumCPU() * int(cfg.GetConcurrencyFactor()))
	i := fromBlock
	j := toBlock

	// Jobs need to be submitted using this interface, while
	// just mentioning which block needs to be fetched
	job := func(num uint64) {
		jd(wp, &d.Job{
			Client: client,
			DB:     _db,
			Redis:  redis,
			Block:  num,
			Status: status,
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
func SyncBlocksByRange(client *ethclient.Client, _db *gorm.DB, redis *data.RedisInfo, fromBlock uint64, toBlock uint64, status *d.StatusHolder) {

	// Job to be submitted and executed by each worker
	//
	// Job specification is provided in `Job` struct
	job := func(wp *workerpool.WorkerPool, j *d.Job) {

		wp.Submit(func() {

			if !HasBlockFinalized(status, j.Block) {

				log.Print(color.Yellow.Sprintf("[x] Finality not yet achieved for block %d [ Latest : %d, In Queue : %d ]", j.Block, status.GetLatestBlockNumber(), getUnfinalizedBlocksQueueLength(redis)))

				// Pushing into unfinalized block queue, to be picked up only when
				// finality for this block has been achieved
				pushBlockNumberIntoUnfinalizedQueue(redis, fmt.Sprintf("%d", j.Block))
				return

			}

			FetchBlockByNumber(j.Client, j.Block, j.DB, j.Redis, j.Status)

		})
	}

	log.Printf("[*] Starting block syncer\n")

	if fromBlock < toBlock {
		Syncer(client, _db, redis, fromBlock, toBlock, status, job)
	} else {
		Syncer(client, _db, redis, toBlock, fromBlock, status, job)
	}

	log.Printf("[+] Stopping block syncer\n")

	// Once completed first iteration of processing blocks upto last time where it left
	// off, we're going to start worker to look at DB & decide which blocks are missing
	// i.e. need to be fetched again
	//
	// And this will itself run as a infinite job, completes one iteration &
	// takes break for 1 min, then repeats
	go SyncMissingBlocksInDB(client, _db, redis, status)
}

// SyncMissingBlocksInDB - Checks with database for what blocks are present & what are not, fetches missing
// blocks & related data iteratively
func SyncMissingBlocksInDB(client *ethclient.Client, _db *gorm.DB, redis *data.RedisInfo, status *d.StatusHolder) {

	// Sleep for 1 minute & then again check whether we need to fetch missing blocks or not
	sleep := func() {
		time.Sleep(time.Duration(1) * time.Minute)
	}

	for {
		log.Printf("[*] Starting missing block finder\n")

		currentBlockNumber := db.GetCurrentBlockNumber(_db)

		// Safely reading shared variable
		blockCount := status.BlockCountInDB()

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

				if !HasBlockFinalized(status, j.Block) {

					log.Print(color.Yellow.Sprintf("[x] Finality not yet achieved for block %d [ Latest : %d, In Queue : %d ]", j.Block, status.GetLatestBlockNumber(), getUnfinalizedBlocksQueueLength(redis)))

					// Pushing into unfinalized block queue, to be picked up only when
					// finality for this block has been achieved
					pushBlockNumberIntoUnfinalizedQueue(redis, fmt.Sprintf("%d", j.Block))
					return

				}

				// Worker fetches block by number from local storage
				block := db.GetBlock(j.DB, j.Block)
				if block == nil && !checkExistenceOfBlockNumberInRedisQueue(redis, fmt.Sprintf("%d", j.Block)) {
					// If not found, block fetching cycle is run, for this block
					FetchBlockByNumber(j.Client, j.Block, j.DB, j.Redis, j.Status)
				}

			})
		}

		Syncer(client, _db, redis, 0, currentBlockNumber, status, job)

		log.Printf("[+] Stopping missing block finder\n")
		sleep()
	}

}
