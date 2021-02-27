package queue

import (
	"sync"
	"time"
)

// BlockProcessorQueue - To be interacted with before attempting to
// process any block
//
// It's concurrent safe
type BlockProcessorQueue struct {
	Blocks map[uint64]*Block
	Lock   *sync.RWMutex
}

// Block - Keeps track of single block i.e. how many
// times attempted till date, last attempted to process
// whether block data has been published on pubsub topic or not,
// is block processing currently
type Block struct {
	Number        uint64
	IsProcessing  bool
	HasPublished  bool
	AttemptCount  uint64
	LastAttempted time.Time
}

// NewQueue - Get a new instance of Block Processor Queue
//
// This needs to be called only single time, application wide
func NewQueue() *BlockProcessorQueue {

	return &BlockProcessorQueue{
		Blocks: make(map[uint64]*Block),
		Lock:   &sync.RWMutex{},
	}

}

// Enqueue - Add new block number in processing queue
//
// Requester go routine is supposed to process this block
// which is why `IsProcessing` state is set to `true`
func (b *BlockProcessorQueue) Enqueue(number uint64) {

	b.Lock.Lock()
	defer b.Lock.Unlock()

	b.Blocks[number] = &Block{
		Number:        number,
		IsProcessing:  true,
		HasPublished:  false,
		AttemptCount:  0,
		LastAttempted: time.Now(),
	}

}
