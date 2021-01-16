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

// Pop oldest block number from Redis queue & try to fetch it in different go routine
//
// Sleeps for 1000 milliseconds
//
// Keeps repeating
func retryBlockFetching(client *ethclient.Client, _db *gorm.DB, redis *data.RedisInfo, status *d.StatusHolder) {
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

		log.Print(color.Cyan.Sprintf("[~] Retrying block : %d [ In Queue : %d ]", parsedBlockNumber, getRetryQueueLength(redis)))

		// Submitting block processor job into pool
		// which will be picked up & processed
		//
		// This will stop us from blindly creating too many go routines
		func(blockNumber uint64) {
			wp.Submit(func() {

				FetchBlockByNumber(client,
					parsedBlockNumber,
					_db,
					redis,
					status)

			})
		}(parsedBlockNumber)
	}
}

// Pushes failed to fetch block number at end of Redis queue
// given it has not already been added
func pushBlockNumberIntoRetryQueue(redis *data.RedisInfo, blockNumber string) {
	// Checking presence first & then deciding whether to add it or not
	if !checkExistenceOfBlockNumberInRetryQueue(redis, blockNumber) {

		if _, err := redis.Client.RPush(context.Background(), redis.BlockRetryQueueName, blockNumber).Result(); err != nil {
			log.Print(color.Red.Sprintf("[!] Failed to push block %s into retry queue : %s", blockNumber, err.Error()))
		}

	}
}

// Checks whether block number is already added in Redis backed retry queue or not
//
// If yes, it'll not be added again
//
// Note: this feature of checking index of value in redis queue,
// was added in Redis v6.0.6 : https://redis.io/commands/lpos
func checkExistenceOfBlockNumberInRetryQueue(redis *data.RedisInfo, blockNumber string) bool {
	if _, err := redis.Client.LPos(context.Background(), redis.BlockRetryQueueName, blockNumber, _redis.LPosArgs{}).Result(); err != nil {
		return false
	}

	return true
}

// Returns redis backed retry queue length
func getRetryQueueLength(redis *data.RedisInfo) int64 {

	blockCount, err := redis.Client.LLen(context.Background(), redis.BlockRetryQueueName).Result()
	if err != nil {
		log.Printf(color.Red.Sprintf("[!] Failed to determine retry queue length : %s", err.Error()))
	}

	return blockCount

}
