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
func (b *BlockProcessorQueue) Enqueue(number uint64) bool {

	// -- First attempt to check whether block is already in queue
	// or not
	// If yes, we don't need to add is again
	// Some go routine will pick it up in sometime future
	b.Lock.RLock()

	if _, ok := b.Blocks[number]; ok {

		b.Lock.RUnlock()
		return false

	}

	b.Lock.RUnlock()
	// -- Done with checking whether block exists or not

	b.Lock.Lock()
	defer b.Lock.Unlock()

	b.Blocks[number] = &Block{
		IsProcessing:  true,
		HasPublished:  false,
		AttemptCount:  0,
		LastAttempted: time.Now(),
	}

	return true

}

// CanPublish - Some go routine might ask queue whether this block's data
// was attempted to be published in some time past or not
//
// If already done, no need to republish data
func (b *BlockProcessorQueue) CanPublish(number uint64) bool {

	b.Lock.RLock()
	defer b.Lock.RUnlock()

	v, ok := b.Blocks[number]
	if !ok {
		return false
	}

	return !v.HasPublished

}
