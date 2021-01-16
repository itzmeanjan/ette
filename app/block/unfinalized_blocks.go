package block

import (
	"context"
	"log"

	_redis "github.com/go-redis/redis/v8"
	"github.com/gookit/color"
	"github.com/itzmeanjan/ette/app/data"
)

// Pushes block number, which has not yet reached finality, into
// Redis backed queue, which will be poped out only when `N` block
// confirmations achieved on top of it
func pushBlockNumberIntoUnfinalizedQueue(redis *data.RedisInfo, blockNumber string) {
	// Checking presence first & then deciding whether to add it or not
	if !checkBlockNumberExistenceInUnfinalizedQueue(redis, blockNumber) {

		if err := redis.Client.RPush(context.Background(), redis.UnfinalizedBlocksQueueName, blockNumber).Err(); err != nil {
			log.Print(color.Red.Sprintf("[!] Failed to push block %s : %s", blockNumber, err.Error()))
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
