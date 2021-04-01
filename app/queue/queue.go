package queue

import (
	"context"
	"sync"
	"time"
)

// Block - Keeps track of single block i.e. how many
// times attempted till date, last attempted to process
// whether block data has been published on pubsub topic or not,
// is block processing currently
type Block struct {
	IsProcessing  bool
	HasPublished  bool
	Done          bool
	AttemptCount  uint64
	LastAttempted time.Time
	Delay         time.Duration
}

// Request - Any request to be placed into
// queue's channels in this form, so that client
// can also receive response/ confirmation over channel
// that they specify
type Request struct {
	BlockNumber  uint64
	ResponseChan chan bool
}

// BlockProcessorQueue - To be interacted with before attempting to
// process any block
//
// It's concurrent safe
type BlockProcessorQueue struct {
	Blocks        map[uint64]*Block
	Lock          *sync.RWMutex
	PutChan       chan Request
	PublishedChan chan Request
	FailedChan    chan Request
	DoneChan      chan Request
}

func (b *BlockProcessorQueue) Start(ctx context.Context) {

	for {
		select {

		case <-ctx.Done():
			return

		case req := <-b.PutChan:

			// Once a block is inserted into processing queue, don't
			// overwrite its history with some new request
			if _, ok := b.Blocks[req.BlockNumber]; !ok {

				req.ResponseChan <- false
				break

			}

			b.Blocks[req.BlockNumber] = &Block{
				IsProcessing:  true,
				HasPublished:  false,
				Done:          false,
				AttemptCount:  1,
				LastAttempted: time.Now().UTC(),
				Delay:         time.Duration(1) * time.Second,
			}
			req.ResponseChan <- true

		case req := <-b.PublishedChan:
			// Worker go rountine marks this block has been
			// published i.e. doesn't denote it has been processed
			// successfully
			//
			// If not, it'll be marked so & no future attempt
			// should try to publish it again over Pub/Sub

			block, ok := b.Blocks[req.BlockNumber]
			if !ok {
				req.ResponseChan <- false
				break
			}

			block.HasPublished = true
			req.ResponseChan <- true

		case req := <-b.FailedChan:

			block, ok := b.Blocks[req.BlockNumber]
			if !ok {
				req.ResponseChan <- false
				break
			}

			block.IsProcessing = false
			block.AttemptCount++
			req.ResponseChan <- true

		case req := <-b.DoneChan:
			// Worker go routine lets us know it has successfully
			// processed block

			block, ok := b.Blocks[req.BlockNumber]
			if !ok {
				req.ResponseChan <- false
				break
			}

			block.IsProcessing = false
			block.Done = true
			req.ResponseChan <- true

		case <-time.After(time.Duration(1000) * time.Millisecond):
			// Do clean up to free up some memory

			buffer := make([]uint64, 0, len(b.Blocks))

			// Finding out which blocks are done processing & we're good to
			// clean those up
			for k, v := range b.Blocks {

				if v.Done {
					buffer = append(buffer, k)
				}

			}

			// Iterative clean up
			for _, v := range buffer {
				delete(b.Blocks, v)
			}

		}
	}

}
