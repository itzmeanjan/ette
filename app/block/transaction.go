package block

import (
	"context"
	"log"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gookit/color"
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

	// Checking whether `ette` is running in real-time data delivery
	// mode or not
	//
	// If yes & this tx, event log data can be published,
	// we'll try to publish data to redis pubsub channel
	// which will be eventually broadcasted to all clients subscribed to
	// topic of their interest
	if cfg.Get("EtteMode") == "2" || cfg.Get("EtteMode") == "3" {
		PublishTx(block.NumberU64(), tx, sender, receipt, redis)
		PublishEvents(block.NumberU64(), receipt, redis)
	}

}
