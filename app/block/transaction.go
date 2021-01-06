package block

import (
	"context"
	"log"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gookit/color"
	c "github.com/itzmeanjan/ette/app/common"
	cfg "github.com/itzmeanjan/ette/app/config"
	d "github.com/itzmeanjan/ette/app/data"
	"github.com/itzmeanjan/ette/app/db"
	"gorm.io/gorm"
)

// FetchTransactionByHash - Fetching specific transaction related data, tries to publish data if required
// & lets listener go routine know about all tx, event data it collected while processing this tx,
// which will be attempted to be stored in database
func FetchTransactionByHash(client *ethclient.Client, block *types.Block, tx *types.Transaction, _db *gorm.DB, redis *d.RedisInfo, publishable bool, _status *d.StatusHolder, returnValChan chan *db.PackedTransaction) {

	receipt, err := client.TransactionReceipt(context.Background(), tx.Hash())
	if err != nil {
		log.Print(color.Red.Sprintf("[!] Failed to fetch tx receipt [ block : %d ] : %s", block.NumberU64(), err.Error()))

		// Passing nil, to denote, failed to fetch all tx data
		// from blockchain node
		returnValChan <- nil
		return
	}

	sender, err := client.TransactionSender(context.Background(), tx, block.Hash(), receipt.TransactionIndex)
	if err != nil {
		log.Print(color.Red.Sprintf("[!] Failed to fetch tx sender [ block : %d ] : %s", block.NumberU64(), err.Error()))

		// Passing nil, to denote, failed to fetch all tx data
		// from blockchain node
		returnValChan <- nil
		return
	}

	// Passing all tx related data to listener go routine
	// so that it can attempt to store whole block data
	// into database
	defer func() {
		returnValChan <- BuildPackedTx(tx, sender, receipt)
	}()

	// This is not a case when real time data is received, rather this is probably
	// a sync attempt to latest state of blockchain
	//
	// So, in this case, we don't need to publish any data on pubsub channel
	if !publishable {
		return
	}

	// Checking whether `ette` is being run in real-time data delivery
	// mode or not
	//
	// If yes & this tx, event log data can be published,
	// we'll try to publish data to redis pubsub channel
	// which will be eventually broadcasted to all clients subscribed to
	// topic of their interest
	if cfg.Get("EtteMode") == "2" || cfg.Get("EtteMode") == "3" {
		PublishTx(block.NumberU64(), tx, sender, receipt, redis)
	}

}

// BuildPackedTx - Putting all information, `ette` will keep for one tx
// into a single structure, so that it becomes easier to pass to & from functions
func BuildPackedTx(tx *types.Transaction, sender common.Address, receipt *types.Receipt) *db.PackedTransaction {

	packedTx := &db.PackedTransaction{}

	if tx.To() == nil {

		packedTx.Tx = db.Transactions{
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

		packedTx.Tx = db.Transactions{
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

	packedTx.Events = make([]*db.Events, len(receipt.Logs))

	for k, v := range receipt.Logs {

		packedTx.Events[k] = &db.Events{
			Origin:          v.Address.Hex(),
			Index:           v.Index,
			Topics:          c.StringifyEventTopics(v.Topics),
			Data:            v.Data,
			TransactionHash: v.TxHash.Hex(),
			BlockHash:       v.BlockHash.Hex(),
		}

	}

	return packedTx

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

	// Publishing event/ log entries to redis pub-sub topic, to be captured by subscribers
	// and sent to client application, who are interested in this piece of data
	// after applying filter
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
