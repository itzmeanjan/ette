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
