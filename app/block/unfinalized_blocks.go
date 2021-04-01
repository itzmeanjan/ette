package block

import (
	"context"
	"fmt"
	"log"
	"strconv"

	_redis "github.com/go-redis/redis/v8"
	"github.com/gookit/color"
	"github.com/itzmeanjan/ette/app/data"
)

// GetOldestBlockFromUnfinalizedQueue - Attempts to read left most element of Redis backed non-final block queue
// i.e. element at 0th index
func GetOldestBlockFromUnfinalizedQueue(redis *data.RedisInfo) string {

	blockNumber, err := redis.Client.LIndex(context.Background(), redis.UnfinalizedBlocksQueue, 0).Result()
	if err != nil {
		return ""
	}

	return blockNumber

}

// CheckIfOldestBlockIsConfirmed - Given oldest block number present in redis backed unfinalized queue
// checks whether this block has yet reached finality or not
func CheckIfOldestBlockIsConfirmed(redis *data.RedisInfo, status *data.StatusHolder) bool {

	oldest := GetOldestBlockFromUnfinalizedQueue(redis)
	if oldest == "" {
		return false
	}

	parsedOldestBlockNumber, err := strconv.ParseUint(oldest, 10, 64)
	if err != nil {
		return false
	}

	return HasBlockFinalized(status, parsedOldestBlockNumber)

}

// PopOldestBlockFromUnfinalizedQueue - Pops oldest block i.e. left most block from non-final block number
// queue, which can be processed now
func PopOldestBlockFromUnfinalizedQueue(redis *data.RedisInfo) uint64 {

	blockNumber, err := redis.Client.LPop(context.Background(), redis.UnfinalizedBlocksQueue).Result()
	if err != nil {
		return 0
	}

	parsedBlockNumber, err := strconv.ParseUint(blockNumber, 10, 64)
	if err != nil {
		return 0
	}

	return parsedBlockNumber

}

// PushBlockIntoUnfinalizedQueue - Pushes block number, which has not yet reached finality, into
// Redis backed queue, which will be poped out only when `N` block
// confirmations achieved on top of it
func PushBlockIntoUnfinalizedQueue(redis *data.RedisInfo, blockNumber string) {
	// Checking presence first & then deciding whether to add it or not
	if !CheckBlockInUnfinalizedQueue(redis, blockNumber) {

		if _, err := redis.Client.RPush(context.Background(), redis.UnfinalizedBlocksQueue, blockNumber).Result(); err != nil {
			log.Print(color.Red.Sprintf("[!] Failed to push block %s into non-final block queue : %s", blockNumber, err.Error()))
		}

	}
}

// MoveUnfinalizedOldestBlockToEnd - Attempts to pop oldest block ( i.e. left most block )
// from unfinalized queue & pushes it back to end of queue, so that other blocks waiting after
// this one can get be attempted to be processed by workers
//
// @note This can be improved using `LMOVE` command of Redis ( >= 6.2.0 )
func MoveUnfinalizedOldestBlockToEnd(redis *data.RedisInfo) {

	PushBlockIntoUnfinalizedQueue(redis, fmt.Sprintf("%d", PopOldestBlockFromUnfinalizedQueue(redis)))

}

// CheckBlockInUnfinalizedQueue - Checks whether block number is already added in
// Redis backed unfinalized queue or not
//
// If yes, it'll not be added again
//
// Note: this feature of checking index of value in redis queue,
// was added in Redis v6.0.6 : https://redis.io/commands/lpos
func CheckBlockInUnfinalizedQueue(redis *data.RedisInfo, blockNumber string) bool {
	if _, err := redis.Client.LPos(context.Background(), redis.UnfinalizedBlocksQueue, blockNumber, _redis.LPosArgs{}).Result(); err != nil {
		return false
	}

	return true
}

// GetUnfinalizedQueueLength - Returns redis backed unfinalized block number queue length
func GetUnfinalizedQueueLength(redis *data.RedisInfo) int64 {

	blockCount, err := redis.Client.LLen(context.Background(), redis.UnfinalizedBlocksQueue).Result()
	if err != nil {
		log.Print(color.Red.Sprintf("[!] Failed to determine non-final block queue length : %s", err.Error()))
	}

	return blockCount

}
