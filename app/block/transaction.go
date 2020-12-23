package block

import (
	"context"
	"log"
	"sync"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-redis/redis/v8"
	c "github.com/itzmeanjan/ette/app/common"
	cfg "github.com/itzmeanjan/ette/app/config"
	d "github.com/itzmeanjan/ette/app/data"
	"github.com/itzmeanjan/ette/app/db"
	"gorm.io/gorm"
)

// Fetching specific transaction related data & persisting in database
func fetchTransactionByHash(client *ethclient.Client, block *types.Block, tx *types.Transaction, _db *gorm.DB, redisClient *redis.Client, redisKey string, publishable bool, _lock *sync.Mutex, _synced *d.SyncState) {
	receipt, err := client.TransactionReceipt(context.Background(), tx.Hash())
	if err != nil {
		// Pushing block number into Redis queue for retrying later
		pushBlockHashIntoRedisQueue(redisClient, redisKey, block.Number().String())

		log.Printf("[!] Failed to fetch tx receipt : %s\n", err.Error())
		return
	}

	// If DB entry already exists for this tx & all events from tx receipt also present
	if res := db.GetTransaction(_db, block.Hash(), tx.Hash()); res != nil && db.CheckPersistanceStatusOfEvents(_db, receipt) {
		return
	}

	sender, err := client.TransactionSender(context.Background(), tx, block.Hash(), receipt.TransactionIndex)
	if err != nil {
		// Pushing block number into Redis queue for retrying later
		pushBlockHashIntoRedisQueue(redisClient, redisKey, block.Number().String())

		log.Printf("[!] Failed to fetch tx sender : %s\n", err.Error())
		return
	}

	if cfg.Get("EtteMode") == "1" || cfg.Get("EtteMode") == "3" {
		db.PutTransaction(_db, tx, receipt, sender)
		db.PutEvent(_db, receipt)
	}

	// This is not a case when real time data is received, rather this is probably
	// a sync attempt to latest state of blockchain
	// So, in this case, we don't need to publish any data on channel
	if !publishable {
		return
	}

	if cfg.Get("EtteMode") == "2" || cfg.Get("EtteMode") == "3" {

		var _publishTx *d.Transaction

		if tx.To() == nil {
			// This is a contract creation tx
			_publishTx = &d.Transaction{
				Hash:      tx.Hash().Hex(),
				From:      sender.Hex(),
				Contract:  receipt.ContractAddress.Hex(),
				Gas:       tx.Gas(),
				GasPrice:  tx.GasPrice().String(),
				Cost:      tx.Cost().String(),
				Nonce:     tx.Nonce(),
				State:     receipt.Status,
				BlockHash: receipt.BlockHash.Hex(),
			}
		} else {
			// This is a normal tx, so we keep contract field empty
			_publishTx = &d.Transaction{
				Hash:      tx.Hash().Hex(),
				From:      sender.Hex(),
				To:        tx.To().Hex(),
				Gas:       tx.Gas(),
				GasPrice:  tx.GasPrice().String(),
				Cost:      tx.Cost().String(),
				Nonce:     tx.Nonce(),
				State:     receipt.Status,
				BlockHash: receipt.BlockHash.Hex(),
			}
		}

		if err := redisClient.Publish(context.Background(), "transaction", _publishTx).Err(); err != nil {
			log.Printf("[!] Failed to publish transaction from block %d : %s\n", block.NumberU64(), err.Error())
		}

		// Publishing event/ log entries to redis pub-sub topic, to be captured by subscribers
		// and sent to client application, who are interested in this piece of data
		// after applying filter
		for _, v := range receipt.Logs {

			if err := redisClient.Publish(context.Background(), "event", &d.Event{
				Origin:          v.Address.Hex(),
				Index:           v.Index,
				Topics:          c.StringifyEventTopics(v.Topics),
				Data:            v.Data,
				TransactionHash: v.TxHash.Hex(),
				BlockHash:       v.BlockHash.Hex(),
			}).Err(); err != nil {
				log.Printf("[!] Failed to publish event from block %d : %s\n", block.NumberU64(), err.Error())
			}

		}

	}

}
