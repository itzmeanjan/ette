package block

import (
	"runtime"
	"sync"

	cfg "github.com/itzmeanjan/ette/app/config"
)

// LockRequest - When some go routines wants to start processing one
// block, they will send a request of this form i.e. invoke `Acquire`
// function with these pieces of data
//
// Where block number is what they want to process & communication channel
// is what to be used for communication between that go routine & lock manager
type LockRequest struct {
	Block         uint64
	Communication chan bool
}

// ProcessQueueLock - Struct for managing info regarding which blocks are
// currently processing, which go routines are waiting for reprocessing
// certain block number, because same block number was received twice
type ProcessQueueLock struct {
	runningQueue     map[uint64]<-chan bool
	runningQueueLock *sync.RWMutex
	waitingQueue     map[uint64]chan<- bool
	done             chan uint64
	wait             chan *LockRequest
}

// Acquire - When ever one go routine is interested in processing
// some block, it'll attempt to acquire lock that block number
//
// If no other block has already acquired lock for that block number
// it'll be allowed to proceed
//
// If some other block is already processing it, it'll be rather sent to
// one waiting queue, where it'll wait until previous job finishes its
// processing
//
// Once that's done, waiting one to be informed over agreed communication
// channel that they're good to proceed
func (p *ProcessQueueLock) Acquire(request *LockRequest) bool {

	_, ok := p.runningQueue[request.Block]
	if !ok {

		// -- Accessing critical section of code, acquiring lock
		p.runningQueueLock.Lock()
		p.runningQueue[request.Block] = request.Communication
		p.runningQueueLock.Unlock()
		// -- Released lock, done with critical section of code

		go p.running(request)
		return true

	}

	p.wait <- request
	return false

}

// Release - When done with processing block, worker invokes method
// to let run tracker know it's done with processing this block, it there's
// any other worker waiting for processing same block, it can now go in
func (p *ProcessQueueLock) Release(comm chan<- bool) {

	comm <- true

}

// NewLock - Create new instance of queue lock manager
// while starting waiting queue watcher in seperate go routine
func NewLock() *ProcessQueueLock {

	lock := &ProcessQueueLock{
		runningQueue:     make(map[uint64]<-chan bool, 0),
		runningQueueLock: &sync.RWMutex{},
		waitingQueue:     make(map[uint64]chan<- bool, 0),
		done:             make(chan uint64, runtime.NumCPU()*int(cfg.GetConcurrencyFactor())),
		wait:             make(chan *LockRequest, runtime.NumCPU()*int(cfg.GetConcurrencyFactor())),
	}

	go lock.waitingQueueWatceher()

	return lock

}

// Looks for new block processing request arriving in waiting queue
// or certain block getting processed & waiting go routines getting notified, that they can
// again attempt to call `Acquire`
//
// This function should be started as one seperate go routine, which will be started when
// processor queue lock manager to be created
func (p *ProcessQueueLock) waitingQueueWatceher() {

	for {

		select {

		case num := <-p.done:

			// As soon as one block processing is done, this go routine will
			// be notified, which will also let any waiting block processor go routine
			// know they're good to go with their processing
			comm, ok := p.waitingQueue[num]
			if ok {

				comm <- true
				delete(p.waitingQueue, num)

			}

		case req := <-p.wait:

			// If there's already one go routine which is waiting for
			// attempt to process same block number, then we're going to
			// ask that block number not to wait anymore
			//
			// because newer processor has arrived, so it'll be prioritized
			//
			// it denotes, for a single block number, there could be at max one
			// worker in running state & one worker in waiting state
			comm, ok := p.waitingQueue[req.Block]
			if ok {

				comm <- false

			}

			p.waitingQueue[req.Block] = req.Communication

		}

	}

}

func (p *ProcessQueueLock) running(request *LockRequest) {

	defer func() {

		// -- Critical section of code, below
		p.runningQueueLock.Lock()
		delete(p.runningQueue, request.Block)
		p.runningQueueLock.Unlock()
		// -- Releasing lock, done with critical section of code

	}()

	// this is a blocking call, only to be unblocked when
	// client has done it's job & call release lock
	<-request.Communication

	// as soon as lock for this block is released,
	// it'll be published to listener, which will check if
	// any party is waiting for this block completion or not
	//
	// if yes, they'll be informed
	p.done <- request.Block
	return

}
