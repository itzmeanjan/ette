package block

import (
	"context"
	"log"
	"runtime"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gammazero/workerpool"
	_redis "github.com/go-redis/redis/v8"
	"github.com/gookit/color"
	cfg "github.com/itzmeanjan/ette/app/config"
	"github.com/itzmeanjan/ette/app/data"
	d "github.com/itzmeanjan/ette/app/data"
	"gorm.io/gorm"
)

// RetryQueueManager - Pop oldest block number from Redis backed retry
// queue & try to fetch it in different go routine
//
// Sleeps for 1000 milliseconds
//
// Keeps repeating
func RetryQueueManager(client *ethclient.Client, _db *gorm.DB, redis *data.RedisInfo, status *d.StatusHolder) {
	sleep := func() {
		time.Sleep(time.Duration(1000) * time.Millisecond)
	}

	// Creating worker pool and submitting jobs as soon as it's determined
	// there's `to be processed` blocks in retry queue
	wp := workerpool.New(runtime.NumCPU() * int(cfg.GetConcurrencyFactor()))
	defer wp.Stop()

	for {
		sleep()

		// Popping oldest element from Redis queue
		blockNumber, err := redis.Client.LPop(context.Background(), redis.BlockRetryQueueName).Result()
		if err != nil {
			continue
		}

		// Parsing string blockNumber to uint64
		parsedBlockNumber, err := strconv.ParseUint(blockNumber, 10, 64)
		if err != nil {
			continue
		}

		log.Print(color.Cyan.Sprintf("[~] Retrying block : %d [ In Queue : %d ]", parsedBlockNumber, GetRetryQueueLength(redis)))

		// Submitting block processor job into pool
		// which will be picked up & processed
		//
		// This will stop us from blindly creating too many go routines
		func(_blockNumber uint64) {

			wp.Submit(func() {

				// This check helps us in determining whether we should
				// consider sending notification over pubsub channel for this block
				// whose processing failed due to some reasons in last attempt
				if status.MaxBlockNumberAtStartUp() <= _blockNumber {

					FetchBlockByNumber(client, _blockNumber, _db, redis, true, status)
					return

				}

				FetchBlockByNumber(client, _blockNumber, _db, redis, false, status)

			})

		}(parsedBlockNumber)
	}
}

// PushBlockIntoRetryQueue - Pushes failed to fetch block number at end of Redis queue
// given it has not already been added
func PushBlockIntoRetryQueue(redis *data.RedisInfo, blockNumber string) {
	// Checking presence first & then deciding whether to add it or not
	if !CheckBlockInRetryQueue(redis, blockNumber) {

		if _, err := redis.Client.RPush(context.Background(), redis.BlockRetryQueueName, blockNumber).Result(); err != nil {
			log.Print(color.Red.Sprintf("[!] Failed to push block %s into retry queue : %s", blockNumber, err.Error()))
		}

	}
}

// CheckBlockInRetryQueue - Checks whether block number is already added in
// Redis backed retry queue or not
//
// If yes, it'll not be added again
//
// Note: this feature of checking index of value in redis queue,
// was added in Redis v6.0.6 : https://redis.io/commands/lpos
func CheckBlockInRetryQueue(redis *data.RedisInfo, blockNumber string) bool {
	if _, err := redis.Client.LPos(context.Background(), redis.BlockRetryQueueName, blockNumber, _redis.LPosArgs{}).Result(); err != nil {
		return false
	}

	return true
}

// GetRetryQueueLength - Returns redis backed retry queue length
func GetRetryQueueLength(redis *data.RedisInfo) int64 {

	blockCount, err := redis.Client.LLen(context.Background(), redis.BlockRetryQueueName).Result()
	if err != nil {
		log.Printf(color.Red.Sprintf("[!] Failed to determine retry queue length : %s", err.Error()))
	}

	return blockCount

}
