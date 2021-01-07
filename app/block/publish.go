package block

import (
	"context"
	"log"

	"github.com/gookit/color"
	d "github.com/itzmeanjan/ette/app/data"
	"github.com/itzmeanjan/ette/app/db"
)

// PublishBlock - Attempts to publish block data to Redis pubsub channel
func PublishBlock(block *db.Blocks, redis *d.RedisInfo) {

	if err := redis.Client.Publish(context.Background(), "block", &d.Block{
		Hash:                block.Hash,
		Number:              block.Number,
		Time:                block.Time,
		ParentHash:          block.ParentHash,
		Difficulty:          block.Difficulty,
		GasUsed:             block.GasUsed,
		GasLimit:            block.GasLimit,
		Nonce:               block.Nonce,
		Miner:               block.Miner,
		Size:                block.Size,
		TransactionRootHash: block.TransactionRootHash,
		ReceiptRootHash:     block.ReceiptRootHash,
	}).Err(); err != nil {
		log.Print(color.Red.Sprintf("[!] Failed to publish block %d in channel : %s", block.Number, err.Error()))
	}

}

// PublishTx - Publishes tx & events in tx, related data to respective
// Redis pubsub channel
func PublishTx(blockNumber uint64, tx *db.Transactions, redis *d.RedisInfo) {

	var pTx *d.Transaction

	if tx.To == "" {
		// This is a contract creation tx
		pTx = &d.Transaction{
			Hash:      tx.Hash,
			From:      tx.From,
			Contract:  tx.Contract,
			Value:     tx.Value,
			Data:      tx.Data,
			Gas:       tx.Gas,
			GasPrice:  tx.GasPrice,
			Cost:      tx.Cost,
			Nonce:     tx.Nonce,
			State:     tx.State,
			BlockHash: tx.BlockHash,
		}
	} else {
		// This is a normal tx, so we keep contract field empty
		pTx = &d.Transaction{
			Hash:      tx.Hash,
			From:      tx.From,
			To:        tx.To,
			Value:     tx.Value,
			Data:      tx.Data,
			Gas:       tx.Gas,
			GasPrice:  tx.GasPrice,
			Cost:      tx.Cost,
			Nonce:     tx.Nonce,
			State:     tx.State,
			BlockHash: tx.BlockHash,
		}
	}

	if err := redis.Client.Publish(context.Background(), "transaction", pTx).Err(); err != nil {
		log.Print(color.Red.Sprintf("[!] Failed to publish transaction from block %d : %s", blockNumber, err.Error()))
	}

}

// PublishEvent - Publishing event/ log entry to redis pub-sub topic, to be captured by subscribers
// and sent to client application, who are interested in this piece of data
// after applying filter
func PublishEvent(blockNumber uint64, event *db.Events, redis *d.RedisInfo) {

	if err := redis.Client.Publish(context.Background(), "event", &d.Event{
		Origin:          event.Origin,
		Index:           event.Index,
		Topics:          event.Topics,
		Data:            event.Data,
		TransactionHash: event.TransactionHash,
		BlockHash:       event.BlockHash,
	}).Err(); err != nil {
		log.Print(color.Red.Sprintf("[!] Failed to publish event from block %d : %s", blockNumber, err.Error()))
	}

}
