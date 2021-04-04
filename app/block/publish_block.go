package block

import (
	"context"
	"log"

	d "github.com/itzmeanjan/ette/app/data"
	"github.com/itzmeanjan/ette/app/db"
)

// PublishBlock - Attempts to publish block data to Redis pubsub channel
func PublishBlock(block *db.PackedBlock, redis *d.RedisInfo) bool {

	if block == nil {
		return false
	}

	_block := &d.Block{
		Hash:                block.Block.Hash,
		Number:              block.Block.Number,
		Time:                block.Block.Time,
		ParentHash:          block.Block.ParentHash,
		Difficulty:          block.Block.Difficulty,
		GasUsed:             block.Block.GasUsed,
		GasLimit:            block.Block.GasLimit,
		Nonce:               block.Block.Nonce,
		Miner:               block.Block.Miner,
		Size:                block.Block.Size,
		StateRootHash:       block.Block.StateRootHash,
		UncleHash:           block.Block.UncleHash,
		TransactionRootHash: block.Block.TransactionRootHash,
		ReceiptRootHash:     block.Block.ReceiptRootHash,
		ExtraData:           block.Block.ExtraData,
	}

	if err := redis.Client.Publish(context.Background(), redis.BlockPublishTopic, _block).Err(); err != nil {

		log.Printf("❗️ Failed to publish block %d : %s\n", block.Block.Number, err.Error())
		return false

	}

	log.Printf("📎 Published block %d\n", block.Block.Number)

	return PublishTxs(block.Block.Number, block.Transactions, redis)

}
