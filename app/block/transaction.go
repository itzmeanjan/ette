package block

import (
	"context"
	"log"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-redis/redis/v8"
	d "github.com/itzmeanjan/ette/app/data"
	"github.com/itzmeanjan/ette/app/db"
	"gorm.io/gorm"
)

// Fetching specific transaction related data & persisting in database
func fetchTransactionByHash(client *ethclient.Client, block *types.Block, tx *types.Transaction, _db *gorm.DB, redisClient *redis.Client) {
	receipt, err := client.TransactionReceipt(context.Background(), tx.Hash())
	if err != nil {
		log.Printf("[!] Failed to fetch tx receipt : %s\n", err.Error())
		return
	}

	// If DB entry already exists for this tx & all events from tx receipt also present
	if res := db.GetTransaction(_db, block.Hash(), tx.Hash()); res != nil && db.CheckPersistanceStatusOfEvents(_db, receipt) {
		return
	}

	sender, err := client.TransactionSender(context.Background(), tx, block.Hash(), receipt.TransactionIndex)
	if err != nil {
		log.Printf("[!] Failed to fetch tx sender : %s\n", err.Error())
		return
	}

	db.PutTransaction(_db, tx, receipt, sender)
	db.PutEvent(_db, receipt)
	log.Printf("[+] Block %d with %d tx(s)\n", block.NumberU64(), len(block.Transactions()))

	// This is not a case when real time data is received, rather this is probably
	// a sync attempt to latest state of blockchain
	// So, in this case, we don't need to publish any data on channel
	if redisClient == nil {
		return
	}

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
}
