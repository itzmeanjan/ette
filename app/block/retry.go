package block

import (
	"context"
	"log"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gammazero/workerpool"
	"github.com/go-redis/redis/v8"
	"github.com/gookit/color"
	d "github.com/itzmeanjan/ette/app/data"
	"gorm.io/gorm"
)

// Pop oldest block number from Redis queue & try to fetch it in different go routine
//
// Sleeps for 1000 milliseconds
//
// Keeps repeating
func retryBlockFetching(client *ethclient.Client, _db *gorm.DB, redisClient *redis.Client, redisKey string, _lock *sync.Mutex, _synced *d.SyncState) {
	sleep := func() {
		time.Sleep(time.Duration(1) * time.Second)
	}

	// Creating worker pool and submitting jobs as soon as it's determined
	// there's `to be processed` blocks in retry queue
	wp := workerpool.New(runtime.NumCPU())
	defer wp.Stop()

	for {
		sleep()

		// Popping oldest element from Redis queue
		blockNumber, err := redisClient.LPop(context.Background(), redisKey).Result()
		if err != nil {
			continue
		}

		// Parsing string blockNumber to uint64
		parsedBlockNumber, err := strconv.ParseUint(blockNumber, 10, 64)
		if err != nil {
			continue
		}

		queuedBlocks, err := redisClient.LLen(context.Background(), redisKey).Result()
		if err != nil {
			log.Printf(color.Red.Sprintf("[!] Failed to determine Redis queue length : %s", err.Error()))
		}

		log.Print(color.Cyan.Sprintf("[~] Retrying block : %d [ In Queue : %d ]", parsedBlockNumber, queuedBlocks))

		// Submitting block processor job into pool
		// which will be picked up & processed
		//
		// This will stop us from blindly creating too many go routines
		func(blockNumber uint64) {
			wp.Submit(func() {

				fetchBlockByNumber(client,
					parsedBlockNumber,
					_db,
					redisClient,
					redisKey,
					_lock,
					_synced)

			})
		}(parsedBlockNumber)
	}
}

// Pushes failed to fetch block number at end of Redis queue
// given it has not already been added
func pushBlockHashIntoRedisQueue(redisClient *redis.Client, redisKey string, blockNumber string) {
	// Checking presence first & then deciding whether to add it or not
	if !checkExistenceOfBlockNumberInRedisQueue(redisClient, redisKey, blockNumber) {

		if err := redisClient.RPush(context.Background(), redisKey, blockNumber).Err(); err != nil {
			log.Print(color.Red.Sprintf("[!] Failed to push block %s : %s", blockNumber, err.Error()))
		}

	}
}

// Checks whether block number is already added in Redis backed retry queue or not
//
// If yes, it'll not be added again
//
// Note: this feature of checking index of value in redis queue,
// was added in Redis v6.0.6 : https://redis.io/commands/lpos
func checkExistenceOfBlockNumberInRedisQueue(redisClient *redis.Client, redisKey string, blockNumber string) bool {
	if _, err := redisClient.LPos(context.Background(), redisKey, blockNumber, redis.LPosArgs{}).Result(); err != nil {
		return false
	}

	return true
}
