package block

import (
	"log"
	"runtime"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gammazero/workerpool"
	"github.com/gookit/color"
	cfg "github.com/itzmeanjan/ette/app/config"
	d "github.com/itzmeanjan/ette/app/data"
	"github.com/itzmeanjan/ette/app/db"
	"gorm.io/gorm"
)

// ProcessBlockContent - Processes everything inside this block i.e. block data, tx data, event data
func ProcessBlockContent(client *ethclient.Client, block *types.Block, _db *gorm.DB, redis *d.RedisInfo, publishable bool, status *d.StatusHolder) {

	if publishable && (cfg.Get("EtteMode") == "1" || cfg.Get("EtteMode") == "3") {
		PublishBlock(block, redis)
	}

	if block.Transactions().Len() == 0 {
		log.Print(color.Green.Sprintf("[+] Block %d with 0 tx(s)", block.NumberU64()))

		status.IncrementBlocksProcessed()
		return
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
		// and expecting their status result to be published on shared channel
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

	// When all tx(s) are successfully processed ( as they have informed us over go channel ),
	// we're happy to exit from this context, given that none of them failed
	if result.Failure == 0 {
		log.Print(color.Green.Sprintf("[+] Block %d with %d tx(s)", block.NumberU64(), len(block.Transactions())))

		status.IncrementBlocksProcessed()
		return
	}

	// Pushing block number into Redis queue for retrying later
	// because it failed to complete some of its jobs 👆
	pushBlockHashIntoRedisQueue(redis, block.Number().String())
}
