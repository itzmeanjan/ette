package queue

import (
	"context"
	"math"
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

// Next - Block to be processed next, asked
// by sending this request & when receptor
// detects so, will attempt to find out
// what should be next processed & send that block
// number is response over channel specified by client
type Next struct {
	ResponseChan chan struct {
		Status bool
		Number uint64
	}
}

// BlockProcessorQueue - To be interacted with before attempting to
// process any block
//
// It's concurrent safe
type BlockProcessorQueue struct {
	Blocks         map[uint64]*Block
	Lock           *sync.RWMutex
	PutChan        chan Request
	CanPublishChan chan Request
	PublishedChan  chan Request
	FailedChan     chan Request
	DoneChan       chan Request
	NextChan       chan Next
}

// Put - Client is supposed to be invoking this method
// when it's interested in putting new block to processing queue
//
// If responded with `true`, they're good to go with execution of
// processing of this block
//
// If this block is already put into queue, it'll ask client
// to not proceed with this number
func (b *BlockProcessorQueue) Put(block uint64) bool {

	resp := make(chan bool)
	req := Request{
		BlockNumber:  block,
		ResponseChan: resp,
	}

	b.PutChan <- req
	return <-resp

}

// CanPublish - Before any client attempts to publish any block
// on Pub/Sub topic, they're supposed to be invoking this method
// to check whether they're eligible of publishing or not
//
// Actually if any other client has already published it, we'll
// better avoid redoing it
func (b *BlockProcessorQueue) CanPublish(block uint64) bool {

	resp := make(chan bool)
	req := Request{
		BlockNumber:  block,
		ResponseChan: resp,
	}

	b.CanPublishChan <- req
	return <-resp

}

// Published - Asks queue manager to mark that this block has been
// successfully published on Pub/Sub topic
//
// Future block processing attempts ( if any ), are supposed to be
// avoiding doing this, if already done successfully
func (b *BlockProcessorQueue) Published(block uint64) bool {

	resp := make(chan bool)
	req := Request{
		BlockNumber:  block,
		ResponseChan: resp,
	}

	b.PublishedChan <- req
	return <-resp

}

func (b *BlockProcessorQueue) Start(ctx context.Context) {

	for {
		select {

		case <-ctx.Done():
			return

		case req := <-b.PutChan:

			// Once a block is inserted into processing queue, don't
			// overwrite its history with some new request
			if _, ok := b.Blocks[req.BlockNumber]; ok {

				req.ResponseChan <- false
				break

			}

			b.Blocks[req.BlockNumber] = &Block{
				IsProcessing:  true,
				HasPublished:  false,
				Done:          false,
				LastAttempted: time.Now().UTC(),
				Delay:         time.Duration(1) * time.Second,
			}
			req.ResponseChan <- true

		case req := <-b.CanPublishChan:

			block, ok := b.Blocks[req.BlockNumber]
			if !ok {
				req.ResponseChan <- false
				break
			}

			req.ResponseChan <- block.HasPublished

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

			// New attempt to process this block can be
			// performed only after current wall time has reached
			// `lastAttempted` + `delay`
			//
			// delay is computed using fibonacci sequence & wrapped
			// at 3600 seconds
			block.Delay = time.Duration(int64(math.Round(block.Delay.Seconds()*(1.0+math.Sqrt(5.0))/2)) % 3600)

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

		case nxt := <-b.NextChan:

			// This is the block number which should be processed
			// by requester client, which is attempted to be found
			var selected uint64
			// Whether we've found anything or not
			var found bool

			for k := range b.Blocks {

				if b.Blocks[k].IsProcessing || b.Blocks[k].Done {
					continue
				}

				if time.Now().UTC().After(b.Blocks[k].LastAttempted.Add(b.Blocks[k].Delay)) {
					selected = k
					found = true

					break
				}

			}

			if !found {

				// As we've failed to find any block which can be processed
				// now, we're asking client to come back sometime later
				//
				// When to come back is upto client
				nxt.ResponseChan <- struct {
					Status bool
					Number uint64
				}{
					Status: false,
				}
				break

			}

			// Updated when last this block was attempted to be processed
			b.Blocks[selected].LastAttempted = time.Now().UTC()
			b.Blocks[selected].IsProcessing = true

			// Asking client to proceed with processing of this block
			nxt.ResponseChan <- struct {
				Status bool
				Number uint64
			}{
				Status: true,
				Number: selected,
			}

		case <-time.After(time.Duration(1000) * time.Millisecond):

			// Finding out which blocks are done processing & we're good to
			// clean those up
			for k := range b.Blocks {

				if b.Blocks[k].Done {
					delete(b.Blocks, k)
				}

			}

		}
	}

}