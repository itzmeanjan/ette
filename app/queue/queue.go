package queue

import (
	"context"
	"math"
	"time"

	"github.com/itzmeanjan/ette/app/config"
)

// Block - Keeps track of single block i.e. how many
// times attempted till date, last attempted to process
// whether block data has been published on pubsub topic or not,
// is block processing currently
type Block struct {
	UnconfirmedProgress bool // 1. Joins
	Published           bool // 2. Pub/Sub publishing
	UnconfirmedDone     bool // 3. Done with processing
	ConfirmedProgress   bool // 4. Attempting confirm whether chain reorg happened or not
	ConfirmedDone       bool // 5. Done with bringing latest changes ✅
	LastAttempted       time.Time
	Delay               time.Duration
}

// SetDelay - Set delay at next fibonacci number in series, interpreted as seconds
func (b *Block) SetDelay() {
	b.Delay = time.Duration(int64(math.Round(b.Delay.Seconds()*(1.0+math.Sqrt(5.0))/2))%3600) * time.Second
}

// ResetDelay - Reset delay back to 1 second
func (b *Block) ResetDelay() {
	b.Delay = time.Duration(1) * time.Second
}

// SetLastAttempted - Updates last attempted to process block
// to current UTC time
func (b *Block) SetLastAttempted() {
	b.LastAttempted = time.Now().UTC()
}

// CanAttempt - Can we attempt to process this block ?
//
// Yes, if waiting phase has elapsed
func (b *Block) CanAttempt() bool {
	return time.Now().UTC().After(b.LastAttempted.Add(b.Delay))
}

// Request - Any request to be placed into
// queue's channels in this form, so that client
// can also receive response/ confirmation over channel
// that they specify
type Request struct {
	BlockNumber  uint64
	ResponseChan chan bool
}

type Update Request

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

// Stat - Clients can query how many blocks present
// in queue currently
type Stat struct {
	ResponseChan chan StatResponse
}

// StatResponse - Statistics of queue to be
// responded back to client in this form
type StatResponse struct {
	UnconfirmedProgress uint64
	UnconfirmedWaiting  uint64
	ConfirmedProgress   uint64
	ConfirmedWaiting    uint64
	Total               uint64
}

// BlockProcessorQueue - To be interacted with before attempting to
// process any block
//
// It's concurrent safe
type BlockProcessorQueue struct {
	Blocks                map[uint64]*Block
	StartedWith           uint64
	TotalInserted         uint64
	LatestBlock           uint64
	Total                 uint64
	PutChan               chan Request
	CanPublishChan        chan Request
	PublishedChan         chan Request
	InsertedChan          chan Request
	UnconfirmedFailedChan chan Request
	UnconfirmedDoneChan   chan Request
	ConfirmedFailedChan   chan Request
	ConfirmedDoneChan     chan Request
	StatChan              chan Stat
	LatestChan            chan Update
	UnconfirmedNextChan   chan Next
	ConfirmedNextChan     chan Next
}

// New - Getting new instance of queue, to be
// invoked during setting up application
func New(startingWith uint64) *BlockProcessorQueue {

	return &BlockProcessorQueue{
		Blocks:                make(map[uint64]*Block),
		StartedWith:           startingWith,
		TotalInserted:         0,
		LatestBlock:           0,
		Total:                 0,
		PutChan:               make(chan Request, 128),
		CanPublishChan:        make(chan Request, 128),
		PublishedChan:         make(chan Request, 128),
		InsertedChan:          make(chan Request, 128),
		UnconfirmedFailedChan: make(chan Request, 128),
		UnconfirmedDoneChan:   make(chan Request, 128),
		ConfirmedFailedChan:   make(chan Request, 128),
		ConfirmedDoneChan:     make(chan Request, 128),
		StatChan:              make(chan Stat, 1),
		LatestChan:            make(chan Update, 1),
		UnconfirmedNextChan:   make(chan Next, 1),
		ConfirmedNextChan:     make(chan Next, 1),
	}

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

// Inserted - Marking this block has been inserted into DB ( not updation, it's insertion )
func (b *BlockProcessorQueue) Inserted(block uint64) bool {

	resp := make(chan bool)
	req := Request{
		BlockNumber:  block,
		ResponseChan: resp,
	}

	b.InsertedChan <- req
	return <-resp

}

// UnconfirmedFailed - Unconfirmed block processing failed
func (b *BlockProcessorQueue) UnconfirmedFailed(block uint64) bool {

	resp := make(chan bool)
	req := Request{
		BlockNumber:  block,
		ResponseChan: resp,
	}

	b.UnconfirmedFailedChan <- req
	return <-resp

}

// UnconfirmedDone - Unconfirmed block processed successfully
func (b *BlockProcessorQueue) UnconfirmedDone(block uint64) bool {

	resp := make(chan bool)
	req := Request{
		BlockNumber:  block,
		ResponseChan: resp,
	}

	b.UnconfirmedDoneChan <- req
	return <-resp

}

// ConfirmedFailed - Confirmed block processing failed
func (b *BlockProcessorQueue) ConfirmedFailed(block uint64) bool {

	resp := make(chan bool)
	req := Request{
		BlockNumber:  block,
		ResponseChan: resp,
	}

	b.ConfirmedFailedChan <- req
	return <-resp

}

// ConfirmedDone - Confirmed block processed successfully
func (b *BlockProcessorQueue) ConfirmedDone(block uint64) bool {

	resp := make(chan bool)
	req := Request{
		BlockNumber:  block,
		ResponseChan: resp,
	}

	b.ConfirmedDoneChan <- req
	return <-resp

}

// Stat - Client's are supposed to be invoking this abstracted method
// for checking queue status
func (b *BlockProcessorQueue) Stat() StatResponse {

	resp := make(chan StatResponse)
	req := Stat{ResponseChan: resp}

	b.StatChan <- req
	return <-resp

}

// Latest - Block head subscriber will update queue manager
// that latest block seen is updated
func (b *BlockProcessorQueue) Latest(num uint64) bool {

	resp := make(chan bool)
	udt := Update{BlockNumber: num, ResponseChan: resp}

	b.LatestChan <- udt
	return <-resp

}

// UnconfirmedNext - Next block that can be processed, present in unconfirmed block queue
func (b *BlockProcessorQueue) UnconfirmedNext() (uint64, bool) {

	resp := make(chan struct {
		Status bool
		Number uint64
	})
	req := Next{ResponseChan: resp}

	b.UnconfirmedNextChan <- req

	v := <-resp
	return v.Number, v.Status

}

// ConfirmedNext - Next block that can be processed, to get confirmation & finalised
func (b *BlockProcessorQueue) ConfirmedNext() (uint64, bool) {

	resp := make(chan struct {
		Status bool
		Number uint64
	})
	req := Next{ResponseChan: resp}

	b.ConfirmedNextChan <- req

	v := <-resp
	return v.Number, v.Status

}

// CanBeConfirmed -Checking whether given block number has reached
// finality as per given user set preference, then it can be attempted
// to be checked again & finally entered into storage
func (b *BlockProcessorQueue) CanBeConfirmed(num uint64) bool {

	if b.LatestBlock < config.GetBlockConfirmations() {
		return false
	}

	return b.LatestBlock-config.GetBlockConfirmations() >= num

}

// Start - You're supposed to be starting this method as an
// independent go routine, with will listen on multiple channels
// & respond back over provided channel ( by client )
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
				UnconfirmedProgress: true,
				LastAttempted:       time.Now().UTC(),
				Delay:               time.Duration(1) * time.Second,
			}
			req.ResponseChan <- true

		case req := <-b.CanPublishChan:

			block, ok := b.Blocks[req.BlockNumber]
			if !ok {
				req.ResponseChan <- false
				break
			}

			req.ResponseChan <- !block.Published

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

			block.Published = true
			req.ResponseChan <- true

		case req := <-b.InsertedChan:
			// Increments how many blocks were inserted into DB

			if _, ok := b.Blocks[req.BlockNumber]; !ok {
				req.ResponseChan <- false
				break
			}

			b.TotalInserted++
			req.ResponseChan <- true

		case req := <-b.UnconfirmedFailedChan:

			block, ok := b.Blocks[req.BlockNumber]
			if !ok {
				req.ResponseChan <- false
				break
			}

			block.UnconfirmedProgress = false
			block.SetDelay()

			req.ResponseChan <- true

		case req := <-b.UnconfirmedDoneChan:

			block, ok := b.Blocks[req.BlockNumber]
			if !ok {
				req.ResponseChan <- false
				break
			}

			block.UnconfirmedProgress = false
			block.UnconfirmedDone = true

			if config.Get("EtteMode") == "HISTORICAL" || config.Get("EtteMode") == "HISTORICAL_AND_REALTIME" {
				block.ConfirmedDone = b.CanBeConfirmed(req.BlockNumber)
			} else {
				block.ConfirmedDone = true // No need to attain this, because we're not putting anything in DB
			}

			block.ResetDelay()
			block.SetLastAttempted()

			req.ResponseChan <- true

		case req := <-b.ConfirmedFailedChan:

			block, ok := b.Blocks[req.BlockNumber]
			if !ok {
				req.ResponseChan <- false
				break
			}

			block.ConfirmedProgress = false
			block.SetDelay()

			req.ResponseChan <- true

		case req := <-b.ConfirmedDoneChan:

			block, ok := b.Blocks[req.BlockNumber]
			if !ok {
				req.ResponseChan <- false
				break
			}

			block.ConfirmedProgress = false
			block.ConfirmedDone = true

			req.ResponseChan <- true

		case nxt := <-b.UnconfirmedNextChan:

			// This is the block number which should be processed
			// by requester client, which is attempted to be found
			var selected uint64
			// Whether we've found anything or not
			var found bool

			for k := range b.Blocks {

				if b.Blocks[k].ConfirmedDone || b.Blocks[k].ConfirmedProgress {
					continue
				}

				if b.Blocks[k].UnconfirmedDone || b.Blocks[k].UnconfirmedProgress {
					continue
				}

				if b.Blocks[k].CanAttempt() {
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
			b.Blocks[selected].SetLastAttempted()
			b.Blocks[selected].UnconfirmedProgress = true

			// Asking client to proceed with processing of this block
			nxt.ResponseChan <- struct {
				Status bool
				Number uint64
			}{
				Status: true,
				Number: selected,
			}

		case nxt := <-b.ConfirmedNextChan:

			var selected uint64
			var found bool

			for k := range b.Blocks {

				if b.Blocks[k].ConfirmedDone || b.Blocks[k].ConfirmedProgress {
					continue
				}

				if !b.Blocks[k].UnconfirmedDone {
					continue
				}

				if b.Blocks[k].CanAttempt() && b.CanBeConfirmed(k) {
					selected = k
					found = true

					break
				}

			}

			if !found {

				nxt.ResponseChan <- struct {
					Status bool
					Number uint64
				}{
					Status: false,
				}
				break

			}

			b.Blocks[selected].SetLastAttempted()
			b.Blocks[selected].ConfirmedProgress = true

			nxt.ResponseChan <- struct {
				Status bool
				Number uint64
			}{
				Status: true,
				Number: selected,
			}

		case req := <-b.StatChan:

			// Returning back how many blocks currently living
			// in block processor queue & in what state
			var stat StatResponse

			for k := range b.Blocks {

				if b.Blocks[k].UnconfirmedProgress {
					stat.UnconfirmedProgress++
					continue
				}

				if b.Blocks[k].UnconfirmedProgress == b.Blocks[k].UnconfirmedDone {
					stat.UnconfirmedWaiting++
					continue
				}

				if b.Blocks[k].ConfirmedProgress {
					stat.ConfirmedProgress++
					continue
				}

				if b.Blocks[k].ConfirmedProgress == b.Blocks[k].ConfirmedDone {
					stat.ConfirmedWaiting++
					continue
				}

			}

			stat.Total = b.Total
			req.ResponseChan <- stat

		case udt := <-b.LatestChan:
			// Latest block number seen by subscriber to
			// sent to queue, to be used in when deciding whether some
			// block is confirmed/ finalised or not
			b.LatestBlock = udt.BlockNumber
			udt.ResponseChan <- true

		case <-time.After(time.Duration(100) * time.Millisecond):

			// Finding out which blocks are confirmed & we're good to
			// clean those up
			for k := range b.Blocks {

				if b.Blocks[k].ConfirmedDone {
					delete(b.Blocks, k)
					b.Total++ // Successfully processed #-of blocks
				}

			}

		}
	}

}
