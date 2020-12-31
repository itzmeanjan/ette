package block

import (
	"context"
	"log"
	"math/big"
	"runtime"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gammazero/workerpool"
	"github.com/go-redis/redis/v8"
	com "github.com/itzmeanjan/ette/app/common"
	d "github.com/itzmeanjan/ette/app/data"
	"github.com/itzmeanjan/ette/app/db"
	"gorm.io/gorm"
)

// Syncer - Given ascending block number range i.e. fromBlock <= toBlock
// fetches blocks in order {fromBlock, toBlock, fromBlock + 1, toBlock - 1, fromBlock + 2, toBlock - 2 ...}
// while running n workers concurrently, where n = number of cores this machine has
//
// Waits for all of them to complete
func Syncer(client *ethclient.Client, _db *gorm.DB, redisClient *redis.Client, redisKey string, fromBlock uint64, toBlock uint64, _lock *sync.Mutex, _synced *d.SyncState, jd func(*workerpool.WorkerPool, *d.Job)) {
	if !(fromBlock <= toBlock) {
		log.Printf("[!] Bad block range for syncer\n")
		return
	}

	wp := workerpool.New(runtime.NumCPU())
	i := fromBlock
	j := toBlock

	// Jobs need to be submitted using this interface, while
	// just mentioning which block needs to be fetched
	job := func(num uint64) {
		jd(wp, &d.Job{
			Client:      client,
			DB:          _db,
			RedisClient: redisClient,
			RedisKey:    redisKey,
			Block:       num,
			Lock:        _lock,
			Synced:      _synced,
		})
	}

	for i <= j {
		// This condition to be arrived at when range has odd number of elements
		if i == j {
			job(i)
		} else {
			job(i)
			job(j)
		}

		i++
		j--
	}

	wp.StopWait()
}

// SyncBlocksByRange - Fetch & persist all blocks in range(fromBlock, toBlock), both inclusive
//
// Range can be either ascending or descending, depending upon that proper arguments to be
// passed to `Syncer` function during invokation
func SyncBlocksByRange(client *ethclient.Client, _db *gorm.DB, redisClient *redis.Client, redisKey string, fromBlock uint64, toBlock uint64, _lock *sync.Mutex, _synced *d.SyncState) {

	// Job to be submitted and executed by each worker
	//
	// Job specification is provided in `Job` struct
	job := func(wp *workerpool.WorkerPool, j *d.Job) {
		wp.Submit(func() {

			fetchBlockByNumber(j.Client, j.Block, j.DB, j.RedisClient, j.RedisKey, j.Lock, j.Synced)

		})
	}

	log.Printf("[*] Starting block syncer\n")

	if fromBlock < toBlock {
		Syncer(client, _db, redisClient, redisKey, fromBlock, toBlock, _lock, _synced, job)
	} else {
		Syncer(client, _db, redisClient, redisKey, toBlock, fromBlock, _lock, _synced, job)
	}

	log.Printf("[+] Stopping block syncer\n")

	// Once completed first iteration of processing blocks upto last time where it left
	// off, we're going to start worker to look at DB & decide which blocks are missing
	// i.e. need to be fetched again
	//
	// And this will itself run as a infinite job, completes one iteration &
	// takes break for 1 min, then repeats
	go SyncMissingBlocksInDB(client, _db, redisClient, redisKey, _lock, _synced)
}

// SyncMissingBlocksInDB - Checks with database for what blocks are present & what are not, fetches missing
// blocks & related data iteratively
func SyncMissingBlocksInDB(client *ethclient.Client, _db *gorm.DB, redisClient *redis.Client, redisKey string, _lock *sync.Mutex, _synced *d.SyncState) {

	// Sleep for 1 minute & then again check whether we need to fetch missing blocks or not
	sleep := func() {
		time.Sleep(time.Duration(1) * time.Minute)
	}

	for {
		log.Printf("[*] Starting missing block finder\n")

		currentBlockNumber := db.GetCurrentBlockNumber(_db)

		// Safely reading shared variable
		_lock.Lock()
		blockCount := _synced.BlockCountInDB
		_lock.Unlock()

		// If all blocks present in between 0 to latest block in network
		// `ette` sleeps for 1 minute & again get to work
		if currentBlockNumber+1 == blockCount {
			log.Printf("[+] No missing blocks found\n")
			sleep()
			continue
		}

		// Job to be submitted and executed by each worker
		//
		// Job specification is provided in `Job` struct
		job := func(wp *workerpool.WorkerPool, j *d.Job) {
			wp.Submit(func() {

				// Worker first fetches block by number from local storage
				block := db.GetBlock(j.DB, j.Block)
				if block == nil {
					// If not found, whole fetching cycle is run, for this block
					fetchBlockByNumber(j.Client, j.Block, j.DB, j.RedisClient, j.RedisKey, j.Lock, j.Synced)
					return
				}

				_num := big.NewInt(0)
				_num = _num.SetUint64(j.Block)

				// Fetch block by number, from Blockchain RPC node
				_fetchedBlock, err := j.Client.BlockByNumber(context.Background(), _num)
				if err != nil {
					log.Printf("[!] Failed to fetch block by number [ block : %d ] : %s\n", j.Block, err)
					return
				}

				// Trying to match what we just fetched from blockchain RPC node and what we've stored in database
				//
				// If similarity not found, we're going to fetch it again
				if !block.SimilarTo(_fetchedBlock) {
					log.Printf("[!] Failed to match with persisted block [ block : %d ] : %s\n", j.Block, err)

					fetchBlockByNumber(j.Client, j.Block, j.DB, j.RedisClient, j.RedisKey, j.Lock, j.Synced)
					return
				}

				// Now trying to go through all tx(s) in this block
				// and checking whether we've all tx and event entries for them
				// or not
				//
				// If not we'll initiate block fetching cycle
				allPresent := true
				for _, t := range _fetchedBlock.Transactions() {

					// Fetching tx receipt from blockchain RPC node
					//
					// If not found, we're going to try a full fetch of this block
					_receipt, err := client.TransactionReceipt(context.Background(), t.Hash())
					if err != nil {
						allPresent = false
						break
					}

					// Fetching tx sender from RPC node
					//
					// If we're failing to fetch so, requesting a full block fetch
					_sender, err := client.TransactionSender(context.Background(), t, _receipt.BlockHash, _receipt.TransactionIndex)
					if err != nil {
						allPresent = false
						break
					}

					// Fetching tx by hash, from local storage
					//
					// If not found, break out of loop and requesting one
					// full block fetch cycle
					_tx := db.GetTransaction(j.DB, t.Hash(), _fetchedBlock.Hash())
					if _tx == nil {
						allPresent = false
						break
					}

					// Checking match between freshly fetched tx data and already persisted tx data
					//
					// If not found, requesting full block fetch
					if !_tx.SimilarTo(t, _receipt, _sender) {
						allPresent = false
						break
					}

					// Iterating over all events present in this tx receipt
					// and checking if any of these not present in database
					allEventsPresent := true
					for _, e := range _receipt.Logs {

						// Fetching event from local storage
						//
						// If not found, one full block fetch cycle is requested
						_event := db.GetEvent(j.DB, e.Index, e.BlockHash)
						if _event == nil {
							allEventsPresent = false
							break
						}

						// Checking similarity between event data we've persisted
						// and what we've just received from blockchain RPC node
						//
						// If not similar, we're going for a full block fetch
						if !_event.SimilarTo(&db.Events{
							Origin:          e.Address.Hex(),
							Index:           e.Index,
							Topics:          com.StringifyEventTopics(e.Topics),
							Data:            e.Data,
							TransactionHash: e.TxHash.Hex(),
							BlockHash:       e.BlockHash.Hex(),
						}) {
							allEventsPresent = false
							break
						}

					}

					// If atleast one of the events not present in database
					// we're going for a full block fetch cycle
					if !allEventsPresent {
						allPresent = false
						break
					}

				}

				// Checking if all tx(s) & event(s) data were present or not
				//
				// If not, we're fetching this block again
				if !allPresent {
					log.Printf("[!] Failed to find all transaction(s)/ event(s) data [ block : %d ]\n", j.Block)

					fetchBlockByNumber(j.Client, j.Block, j.DB, j.RedisClient, j.RedisKey, j.Lock, j.Synced)
					return
				}

			})
		}

		Syncer(client, _db, redisClient, redisKey, 0, currentBlockNumber, _lock, _synced, job)

		log.Printf("[+] Stopping missing block finder\n")
		sleep()
	}

}
