package data

import (
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

// SyncState - Whether `ette` is synced with blockchain or not
type SyncState struct {
	Done                    uint64
	StartedAt               time.Time
	BlockCountAtStartUp     uint64
	MaxBlockNumberAtStartUp uint64
	NewBlocksInserted       uint64
	LatestBlockNumber       uint64
}

// BlockCountInDB - Blocks currently present in database
func (s *SyncState) BlockCountInDB() uint64 {
	return s.BlockCountAtStartUp + s.NewBlocksInserted
}

// StatusHolder - Keeps track of progress being made by `ette` over time,
// which is to be delivered when `/v1/synced` is queried
type StatusHolder struct {
	State *SyncState
	Mutex *sync.RWMutex
}

// MaxBlockNumberAtStartUp - Attempting to safely read latest block number
// when `ette` was started, will help us in deciding whether a missing
// block related notification needs to be sent on a pubsub channel or not
func (s *StatusHolder) MaxBlockNumberAtStartUp() uint64 {

	s.Mutex.RLock()
	defer s.Mutex.RUnlock()

	return s.State.MaxBlockNumberAtStartUp

}

// SetStartedAt - Sets started at time
func (s *StatusHolder) SetStartedAt() {

	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	s.State.StartedAt = time.Now().UTC()

}

// IncrementBlocksInserted - Increments number of blocks inserted into DB
// after `ette` started processing blocks
func (s *StatusHolder) IncrementBlocksInserted() {

	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	s.State.NewBlocksInserted++

}

// IncrementBlocksProcessed - Increments number of blocks processed by `ette
// after it started
func (s *StatusHolder) IncrementBlocksProcessed() {

	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	s.State.Done++

}

// BlockCountInDB - Safely reads currently present blocks in database
func (s *StatusHolder) BlockCountInDB() uint64 {

	s.Mutex.RLock()
	defer s.Mutex.RUnlock()

	return s.State.BlockCountInDB()

}

// ElapsedTime - Uptime of `ette`
func (s *StatusHolder) ElapsedTime() time.Duration {

	s.Mutex.RLock()
	defer s.Mutex.RUnlock()

	return time.Now().UTC().Sub(s.State.StartedAt)

}

// Done - #-of Blocks processed during `ette` uptime i.e. after last time it started
func (s *StatusHolder) Done() uint64 {

	s.Mutex.RLock()
	defer s.Mutex.RUnlock()

	return s.State.Done

}

// GetLatestBlockNumber - Attempting to safely read latest block number seen
func (s *StatusHolder) GetLatestBlockNumber() uint64 {

	s.Mutex.RLock()
	defer s.Mutex.RUnlock()

	return s.State.LatestBlockNumber

}

// SetLatestBlockNumber - Attempting to safely write latest block number
func (s *StatusHolder) SetLatestBlockNumber(num uint64) {

	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	s.State.LatestBlockNumber = num

}

// RedisInfo - Holds redis related information in this struct, to be used
// when passing to functions as argument
type RedisInfo struct {
	Client                                               *redis.Client // using this object `ette` will talk to Redis
	BlockPublishTopic, TxPublishTopic, EventPublishTopic string
}

// ResultStatus - Keeps track of how many operations went successful
// and how many of them failed
type ResultStatus struct {
	Success uint64
	Failure uint64
}

// Total - Returns total count of operations which were supposed to be
// performed
//
// To be useful when deciding whether all go routines have sent their status i.e. completed
// their task or not
func (r ResultStatus) Total() uint64 {
	return r.Success + r.Failure
}

// Job - For running a block fetching job, these are all the information which are required
type Job struct {
	Client *ethclient.Client
	DB     *gorm.DB
	Redis  *RedisInfo
	Block  uint64
	Status *StatusHolder
}

// BlockChainNodeConnection - Holds network connection object for blockchain nodes
//
// Use `RPC` i.e. HTTP based connection, for querying blockchain for data
// Use `Websocket` for real-time listening of events in blockchain
type BlockChainNodeConnection struct {
	RPC       *ethclient.Client
	Websocket *ethclient.Client
}
