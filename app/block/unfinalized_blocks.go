package block

import (
	"context"
	"log"
	"strconv"

	_redis "github.com/go-redis/redis/v8"
	"github.com/gookit/color"
	"github.com/itzmeanjan/ette/app/data"
)

// Attempts to read left most element of Redis backed non-final block queue
// i.e. element at 0th index
func getOldestBlockNumberFromUnfinalizedQueue(redis *data.RedisInfo) string {

	blockNumber, err := redis.Client.LIndex(context.Background(), redis.UnfinalizedBlocksQueueName, 0).Result()
	if err != nil {
		return ""
	}

	return blockNumber

}

// Given oldest block number present in redis backed unfinalized queue
// checks whether this block has yet reached finality or not
func checkIfOldestBlockNumberIsConfirmed(redis *data.RedisInfo, status *data.StatusHolder) bool {

	oldest := getOldestBlockNumberFromUnfinalizedQueue(redis)
	if oldest == "" {
		return false
	}

	parsedOldestBlockNumber, err := strconv.ParseUint(oldest, 10, 64)
	if err != nil {
		return false
	}

	return HasBlockFinalized(status, parsedOldestBlockNumber)

}

// Pops oldest block i.e. left most block from non-final block number
// queue, which can be processed now
func popOldestBlockNumberFromUnfinalizedQueue(redis *data.RedisInfo) uint64 {

	blockNumber, err := redis.Client.LPop(context.Background(), redis.UnfinalizedBlocksQueueName).Result()
	if err != nil {
		return 0
	}

	parsedBlockNumber, err := strconv.ParseUint(blockNumber, 10, 64)
	if err != nil {
		return 0
	}

	return parsedBlockNumber

}

// Pushes block number, which has not yet reached finality, into
// Redis backed queue, which will be poped out only when `N` block
// confirmations achieved on top of it
func pushBlockNumberIntoUnfinalizedQueue(redis *data.RedisInfo, blockNumber string) {
	// Checking presence first & then deciding whether to add it or not
	if !checkBlockNumberExistenceInUnfinalizedQueue(redis, blockNumber) {

		if _, err := redis.Client.RPush(context.Background(), redis.UnfinalizedBlocksQueueName, blockNumber).Result(); err != nil {
			log.Print(color.Red.Sprintf("[!] Failed to push block %s into non-final block queue : %s", blockNumber, err.Error()))
		}

	}
}

// Checks whether block number is already added in Redis backed unfinalized queue or not
//
// If yes, it'll not be added again
//
// Note: this feature of checking index of value in redis queue,
// was added in Redis v6.0.6 : https://redis.io/commands/lpos
func checkBlockNumberExistenceInUnfinalizedQueue(redis *data.RedisInfo, blockNumber string) bool {
	if _, err := redis.Client.LPos(context.Background(), redis.UnfinalizedBlocksQueueName, blockNumber, _redis.LPosArgs{}).Result(); err != nil {
		return false
	}

	return true
}

// Returns redis backed unfinalized block number queue length
func getUnfinalizedBlocksQueueLength(redis *data.RedisInfo) int64 {

	blockCount, err := redis.Client.LLen(context.Background(), redis.UnfinalizedBlocksQueueName).Result()
	if err != nil {
		log.Printf(color.Red.Sprintf("[!] Failed to determine non-final block queue length : %s", err.Error()))
	}

	return blockCount

}
