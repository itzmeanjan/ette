package block

import (
	"context"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-redis/redis/v8"
	d "github.com/itzmeanjan/ette/app/data"
	"gorm.io/gorm"
)

// Pop oldest block hash from Redis queue & try to fetch it in different go routine
//
// Sleeps for 500 milliseconds
//
// Keeps repeating
func retryBlockFetching(client *ethclient.Client, _db *gorm.DB, redisClient *redis.Client, redisKey string, _lock *sync.Mutex, _synced *d.SyncState) {
	sleep := func() {
		time.Sleep(time.Duration(500) * time.Millisecond)
	}

	for {

		// Popping oldest element from Redis queue
		blockNumber, err := redisClient.LPop(context.Background(), redisKey).Result()
		if err != nil {
			sleep()
		}

		// Parsing string blockNumber to uint64
		parsedBlockNumber, err := strconv.ParseUint(blockNumber, 10, 64)
		if err != nil {
			sleep()
		}

		log.Printf("[~] Retrying block : %d\n", parsedBlockNumber)
		go fetchBlockByNumber(client, parsedBlockNumber, _db, redisClient, redisKey, _lock, _synced)
		sleep()
	}
}

// Pushes failed to fetch block hash at end of Redis queue
func pushBlockHashIntoRedisQueue(redisClient *redis.Client, redisKey string, blockNumber string) {
	if err := redisClient.RPush(context.Background(), redisKey, blockNumber).Err(); err != nil {
		log.Printf("[!] Failed to push block %s : %s\n", blockNumber, err.Error())
	}
}
