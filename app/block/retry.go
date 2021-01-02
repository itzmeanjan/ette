package block

import (
	"context"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-redis/redis/v8"
	"github.com/gookit/color"
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
		go fetchBlockByNumber(client, parsedBlockNumber, _db, redisClient, redisKey, _lock, _synced)
	}
}

// Pushes failed to fetch block hash at end of Redis queue
func pushBlockHashIntoRedisQueue(redisClient *redis.Client, redisKey string, blockNumber string) {
	if err := redisClient.RPush(context.Background(), redisKey, blockNumber).Err(); err != nil {
		log.Print(color.Red.Sprintf("[!] Failed to push block %s : %s", blockNumber, err.Error()))
	}
}
