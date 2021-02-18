package block

import (
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gammazero/workerpool"
	"github.com/gookit/color"
	cfg "github.com/itzmeanjan/ette/app/config"
	d "github.com/itzmeanjan/ette/app/data"
	"github.com/itzmeanjan/ette/app/db"
	"gorm.io/gorm"
)

// HasBlockFinalized - Checking whether block under processing i.e. `number`
// has `N` confirmations on top of it or not
func HasBlockFinalized(status *d.StatusHolder, number uint64) bool {

	return status.GetLatestBlockNumber()-cfg.GetBlockConfirmations() >= number

}

// ProcessBlockContent - Processes everything inside this block i.e. block data, tx data, event data
func ProcessBlockContent(client *ethclient.Client, block *types.Block, _db *gorm.DB, redis *d.RedisInfo, publishable bool, status *d.StatusHolder, startingAt time.Time) bool {

	// Closure managing publishing whole block data i.e. block header, txn(s), event logs
	// on redis pubsub channel
	pubsubWorker := func(txns []*db.PackedTransaction) *db.PackedBlock {

		// Constructing block data to published & persisted
		packedBlock := BuildPackedBlock(block, txns)

		// Attempting to publish whole block data to redis pubsub channel
		// when eligible `EtteMode` is set
		if publishable && (cfg.Get("EtteMode") == "2" || cfg.Get("EtteMode") == "3") {
			PublishBlock(packedBlock, redis)
		}

		return packedBlock

	}

	if block.Transactions().Len() == 0 {

		// Constructing block data to be persisted
		//
		// This is what we just published on pubsub channel
		packedBlock := pubsubWorker(nil)

		// If `ette` being run in mode, for only publishing data to
		// pubsub channel, no need to persist data
		//
		// We simply publish & return from execution scope
		if !(cfg.Get("EtteMode") == "1" || cfg.Get("EtteMode") == "3") {

			log.Print(color.Green.Sprintf("[+] Block %d with 0 tx(s) [ Took : %s ]", block.NumberU64(), time.Now().UTC().Sub(startingAt)))
			status.IncrementBlocksProcessed()

			return true

		}

		// If block doesn't contain any tx, we'll attempt to persist only block
		if err := db.StoreBlock(_db, packedBlock, status); err != nil {

			log.Print(color.Red.Sprintf("[+] Failed to process block %d with 0 tx(s) : %s [ Took : %s ]", block.NumberU64(), err.Error(), time.Now().UTC().Sub(startingAt)))

			// If failed to persist, we'll put it in retry queue
			PushBlockIntoRetryQueue(redis, block.Number().String())
			return false

		}

		if !HasBlockFinalized(status, packedBlock.Block.Number) {

			log.Print(color.LightRed.Sprintf("[x] Non-final block %d with 0 tx(s) [ Took : %s | Latest Block : %d | In Queue : %d ]", packedBlock.Block.Number, time.Now().UTC().Sub(startingAt), status.GetLatestBlockNumber(), GetUnfinalizedQueueLength(redis)))

			// Pushing into unfinalized block queue, to be picked up only when
			// finality for this block has been achieved
			PushBlockIntoUnfinalizedQueue(redis, fmt.Sprintf("%d", packedBlock.Block.Number))
			return true

		}

		// Successfully processed block
		log.Print(color.Green.Sprintf("[+] Block %d with 0 tx(s) [ Took : %s ]", block.NumberU64(), time.Now().UTC().Sub(startingAt)))
		status.IncrementBlocksProcessed()

		return true

	}

	// Communication channel to be shared between multiple executing go routines
	// which are trying to fetch all tx(s) present in block, concurrently
	returnValChan := make(chan *db.PackedTransaction, runtime.NumCPU()*int(cfg.GetConcurrencyFactor()))

	// -- Tx processing starting
	// Creating job processor queue
	// which will process all tx(s), concurrently
	wp := workerpool.New(runtime.NumCPU() * int(cfg.GetConcurrencyFactor()))

	// Concurrently trying to process all tx(s) for this block, in hope of better performance
	for _, v := range block.Transactions() {

		// Concurrently trying to fetch multiple tx(s) present in block
		// and expecting their return value to be published on shared channel
		//
		// Which is being read 👇
		func(tx *types.Transaction) {
			wp.Submit(func() {

				FetchTransactionByHash(client,
					block,
					tx,
					_db,
					redis,
					publishable,
					status,
					returnValChan)

			})
		}(v)

	}

	// Keeping track of how many of these tx fetchers succeded & how many of them failed
	result := d.ResultStatus{}
	// Data received from tx fetchers, to be stored here
	packedTxs := make([]*db.PackedTransaction, block.Transactions().Len())

	for v := range returnValChan {
		if v != nil {
			result.Success++
		} else {
			result.Failure++
		}

		// #-of tx fetchers completed their job till now
		//
		// Either successfully or failed some how
		total := int(result.Total())
		// Storing tx data received from just completed go routine
		packedTxs[total-1] = v

		// All go routines have completed their job
		if total == block.Transactions().Len() {
			break
		}
	}

	// Stopping job processor forcefully
	// because by this time all jobs have been completed
	//
	// Otherwise control flow will not be able to come here
	// it'll keep looping in 👆 loop, reading from channel
	wp.Stop()
	// -- Tx processing ending

	// When all tx(s) aren't successfully processed ( as they have informed us over go channel ),
	// we're exiting from this context, while putting this block number in retry queue
	if !(result.Failure == 0) {

		PushBlockIntoRetryQueue(redis, block.Number().String())
		return false

	}

	// Constructing block data to be persisted
	//
	// This is what we just published on pubsub channel
	packedBlock := pubsubWorker(packedTxs)

	// If `ette` being run in mode, for only publishing data to
	// pubsub channel, no need to persist data
	//
	// We simply publish & return from execution scope
	if !(cfg.Get("EtteMode") == "1" || cfg.Get("EtteMode") == "3") {

		log.Print(color.Green.Sprintf("[+] Block %d with %d tx(s) [ Took : %s ]", block.NumberU64(), block.Transactions().Len(), time.Now().UTC().Sub(startingAt)))
		status.IncrementBlocksProcessed()

		return true

	}

	// If block doesn't contain any tx, we'll attempt to persist only block
	if err := db.StoreBlock(_db, packedBlock, status); err != nil {

		log.Print(color.Red.Sprintf("[+] Failed to process block %d with %d tx(s) : %s [ Took : %s ]", block.NumberU64(), block.Transactions().Len(), err.Error(), time.Now().UTC().Sub(startingAt)))

		// If failed to persist, we'll put it in retry queue
		PushBlockIntoRetryQueue(redis, block.Number().String())
		return false

	}

	if !HasBlockFinalized(status, packedBlock.Block.Number) {

		log.Print(color.LightRed.Sprintf("[x] Non-final block %d with %d tx(s) [ Took : %s | Latest Block : %d | In Queue : %d ]", packedBlock.Block.Number, block.Transactions().Len(), time.Now().UTC().Sub(startingAt), status.GetLatestBlockNumber(), GetUnfinalizedQueueLength(redis)))

		// Pushing into unfinalized block queue, to be picked up only when
		// finality for this block has been achieved
		PushBlockIntoUnfinalizedQueue(redis, fmt.Sprintf("%d", packedBlock.Block.Number))
		return true

	}

	// Successfully processed block
	log.Print(color.Green.Sprintf("[+] Block %d with %d tx(s) [ Took : %s ]", block.NumberU64(), block.Transactions().Len(), time.Now().UTC().Sub(startingAt)))

	status.IncrementBlocksProcessed()
	return true

}
