package block

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
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
func retryBlockFetching(client *ethclient.Client, _db *gorm.DB, redisClient *redis.Client, redisKeyName string, _lock *sync.Mutex, _synced *d.SyncState) {
	sleep := func() {
		time.Sleep(time.Duration(500) * time.Millisecond)
	}

	for {

		blockHash, err := redisClient.LPop(context.Background(), redisKeyName).Result()
		if !(err == nil && len(blockHash) == 66) {
			sleep()
		}

		log.Printf("[~] Retrying block : %s\n", blockHash)
		go fetchBlockByHash(client, common.HexToHash(blockHash), _db, redisClient, _lock, _synced)
		sleep()
	}
}

// Pushes failed to fetch block hash at end of Redis queue
func pushBlockHashIntoRedisQueue(redisClient *redis.Client, redisKeyName string, blockHash common.Hash) {
	if err := redisClient.RPush(context.Background(), redisKeyName, blockHash.Hex()).Err(); err != nil {
		log.Printf("[!] Failed to push block : %s\n", blockHash.Hex())
	}
}
