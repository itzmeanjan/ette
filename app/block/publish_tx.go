package block

import (
	"context"
	"log"

	"github.com/gookit/color"
	d "github.com/itzmeanjan/ette/app/data"
	"github.com/itzmeanjan/ette/app/db"
)

// PublishTxs - Publishes all transactions in a block to redis pubsub
// channel
func PublishTxs(blockNumber uint64, txs []*db.PackedTransaction, redis *d.RedisInfo) bool {

	if txs == nil {
		return false
	}

	var eventCount uint64
	var status bool

	for _, t := range txs {

		status = PublishTx(blockNumber, t, redis)
		if !status {
			break
		}

		// how many events are present in this block, in total
		eventCount += uint64(len(t.Events))

	}

	log.Print(color.LightMagenta.Sprintf("[*] Published %d transactions of block %d", len(txs), blockNumber))
	log.Print(color.LightMagenta.Sprintf("[*] Published %d events of block %d", eventCount, blockNumber))

	return status

}

// PublishTx - Publishes tx & events in tx, related data to respective
// Redis pubsub channel
func PublishTx(blockNumber uint64, tx *db.PackedTransaction, redis *d.RedisInfo) bool {

	if tx == nil {
		return false
	}

	var pTx *d.Transaction

	if tx.Tx.To == "" {
		// This is a contract creation tx
		pTx = &d.Transaction{
			Hash:      tx.Tx.Hash,
			From:      tx.Tx.From,
			Contract:  tx.Tx.Contract,
			Value:     tx.Tx.Value,
			Data:      tx.Tx.Data,
			Gas:       tx.Tx.Gas,
			GasPrice:  tx.Tx.GasPrice,
			Cost:      tx.Tx.Cost,
			Nonce:     tx.Tx.Nonce,
			State:     tx.Tx.State,
			BlockHash: tx.Tx.BlockHash,
		}
	} else {
		// This is a normal tx, so we keep contract field empty
		pTx = &d.Transaction{
			Hash:      tx.Tx.Hash,
			From:      tx.Tx.From,
			To:        tx.Tx.To,
			Value:     tx.Tx.Value,
			Data:      tx.Tx.Data,
			Gas:       tx.Tx.Gas,
			GasPrice:  tx.Tx.GasPrice,
			Cost:      tx.Tx.Cost,
			Nonce:     tx.Tx.Nonce,
			State:     tx.Tx.State,
			BlockHash: tx.Tx.BlockHash,
		}
	}

	if err := redis.Client.Publish(context.Background(), "transaction", pTx).Err(); err != nil {

		log.Print(color.Red.Sprintf("[!] Failed to publish transaction from block %d : %s", blockNumber, err.Error()))
		return false

	}

	return PublishEvents(blockNumber, tx.Events, redis)

}
