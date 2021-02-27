package queue

import (
	"errors"
	"fmt"
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
		LastAttempted: time.Now().UTC(),
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

// SetPublished - When go routine has completed publishing
// it on pubsub topic, it can mark this block published so
// that even if this block is not completely processed in this
// attempt, in next iteration it'll not be published again
//
// This function is supposed to prevent `> 1 time publishing`
// of same block, scenario
func (b *BlockProcessorQueue) SetPublished(number uint64) error {

	b.Lock.Lock()
	defer b.Lock.Unlock()

	v, ok := b.Blocks[number]
	if !ok {

		return fmt.Errorf("Expected block %d to exist in queue", number)

	}

	v.HasPublished = true
	return nil

}

// SetFailed - When one block was attempted to be processed, but
// failed to complete, it'll be marked that caller go routine is
// not processing it now anymore & failed attempt count to be
// incremented by 1
func (b *BlockProcessorQueue) SetFailed(number uint64) error {

	b.Lock.Lock()
	defer b.Lock.Unlock()

	v, ok := b.Blocks[number]
	if !ok {

		return fmt.Errorf("Expected block %d to exist in queue", number)

	}

	v.AttemptCount++
	v.IsProcessing = false

	return nil

}

// Done - When a block has been successfully processed
// it'll be deleted from entry table
func (b *BlockProcessorQueue) Done(number uint64) error {

	b.Lock.Lock()
	defer b.Lock.Unlock()

	if _, ok := b.Blocks[number]; !ok {

		return fmt.Errorf("Expected block %d to exist in queue", number)

	}

	delete(b.Blocks, number)

	return nil

}

// Next - Block processor go routine asks for next block it can process
// and block which was attempted to be processed longest time ago
// will be prioritized
func (b *BlockProcessorQueue) Next() (uint64, error) {

	b.Lock.RLock()
	defer b.Lock.RUnlock()

	if len(b.Blocks) == 0 {

		return 0, errors.New("Nothing in queue")

	}

	var number uint64
	var oldest time.Time = time.Now().UTC()

	for k, v := range b.Blocks {

		if !(v.AttemptCount > 0) {
			continue
		}

		if v.IsProcessing {
			continue
		}

		if oldest.Before(v.LastAttempted) {

			number = k
			oldest = v.LastAttempted

		}

	}

	return number, nil

}
