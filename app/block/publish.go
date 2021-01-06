package block

import (
	"context"
	"log"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/gookit/color"
	c "github.com/itzmeanjan/ette/app/common"
	d "github.com/itzmeanjan/ette/app/data"
)

// PublishBlock - Attempts to publish block data to Redis pubsub channel
func PublishBlock(block *types.Block, redis *d.RedisInfo) {

	if err := redis.Client.Publish(context.Background(), "block", &d.Block{
		Hash:                block.Hash().Hex(),
		Number:              block.NumberU64(),
		Time:                block.Time(),
		ParentHash:          block.ParentHash().Hex(),
		Difficulty:          block.Difficulty().String(),
		GasUsed:             block.GasUsed(),
		GasLimit:            block.GasLimit(),
		Nonce:               block.Nonce(),
		Miner:               block.Coinbase().Hex(),
		Size:                float64(block.Size()),
		TransactionRootHash: block.TxHash().Hex(),
		ReceiptRootHash:     block.ReceiptHash().Hex(),
	}).Err(); err != nil {
		log.Print(color.Red.Sprintf("[!] Failed to publish block %d in channel : %s", block.NumberU64(), err.Error()))
	}

}

// PublishTx - Publishes tx & events in tx, related data to respective
// Redis pubsub channel
func PublishTx(blockNumber uint64, tx *types.Transaction, sender common.Address, receipt *types.Receipt, redis *d.RedisInfo) {

	var pTx *d.Transaction

	if tx.To() == nil {
		// This is a contract creation tx
		pTx = &d.Transaction{
			Hash:      tx.Hash().Hex(),
			From:      sender.Hex(),
			Contract:  receipt.ContractAddress.Hex(),
			Value:     tx.Value().String(),
			Data:      tx.Data(),
			Gas:       tx.Gas(),
			GasPrice:  tx.GasPrice().String(),
			Cost:      tx.Cost().String(),
			Nonce:     tx.Nonce(),
			State:     receipt.Status,
			BlockHash: receipt.BlockHash.Hex(),
		}
	} else {
		// This is a normal tx, so we keep contract field empty
		pTx = &d.Transaction{
			Hash:      tx.Hash().Hex(),
			From:      sender.Hex(),
			To:        tx.To().Hex(),
			Value:     tx.Value().String(),
			Data:      tx.Data(),
			Gas:       tx.Gas(),
			GasPrice:  tx.GasPrice().String(),
			Cost:      tx.Cost().String(),
			Nonce:     tx.Nonce(),
			State:     receipt.Status,
			BlockHash: receipt.BlockHash.Hex(),
		}
	}

	if err := redis.Client.Publish(context.Background(), "transaction", pTx).Err(); err != nil {
		log.Print(color.Red.Sprintf("[!] Failed to publish transaction from block %d : %s", blockNumber, err.Error()))
	}

}

// PublishEvents - Publishing event/ log entries to redis pub-sub topic, to be captured by subscribers
// and sent to client application, who are interested in this piece of data
// after applying filter
func PublishEvents(blockNumber uint64, receipt *types.Receipt, redis *d.RedisInfo) {

	for _, v := range receipt.Logs {

		if err := redis.Client.Publish(context.Background(), "event", &d.Event{
			Origin:          v.Address.Hex(),
			Index:           v.Index,
			Topics:          c.StringifyEventTopics(v.Topics),
			Data:            v.Data,
			TransactionHash: v.TxHash.Hex(),
			BlockHash:       v.BlockHash.Hex(),
		}).Err(); err != nil {
			log.Print(color.Red.Sprintf("[!] Failed to publish event from block %d : %s", blockNumber, err.Error()))
		}

	}

}
