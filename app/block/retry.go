package block

import (
	"log"
	"runtime"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gammazero/workerpool"
	"github.com/gookit/color"
	cfg "github.com/itzmeanjan/ette/app/config"
	d "github.com/itzmeanjan/ette/app/data"
	q "github.com/itzmeanjan/ette/app/queue"
	"gorm.io/gorm"
)

// RetryQueueManager - Pop oldest block number from Redis backed retry
// queue & try to fetch it in different go routine
//
// Sleeps for 1000 milliseconds
//
// Keeps repeating
func RetryQueueManager(client *ethclient.Client, _db *gorm.DB, redis *d.RedisInfo, queue *q.BlockProcessorQueue, status *d.StatusHolder) {
	sleep := func() {
		time.Sleep(time.Duration(512) * time.Millisecond)
	}

	// Creating worker pool and submitting jobs as soon as it's determined
	// there's `to be processed` blocks in retry queue
	wp := workerpool.New(runtime.NumCPU() * int(cfg.GetConcurrencyFactor()))
	defer wp.Stop()

	for {
		sleep()

		block, ok := queue.Next()
		if !ok {
			continue
		}

		stat := queue.Stat()
		log.Print(color.Cyan.Sprintf("[~] Retrying block : %d [ In Progress : %d, Waiting : %d, Done : %d ]", block, stat.InProgress, stat.Waiting, stat.Done))

		// Submitting block processor job into pool
		// which will be picked up & processed
		//
		// This will stop us from blindly creating too many go routines
		func(_blockNumber uint64, queue *q.BlockProcessorQueue) {

			wp.Submit(func() {

				if !FetchBlockByNumber(client, _blockNumber, _db, redis, true, queue, status) {

					queue.Failed(_blockNumber)
					return

				}

				queue.Done(_blockNumber)

			})

		}(block, queue)
	}
}
