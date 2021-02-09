package block

import "sync"

// ProcessQueueLock - ...
type ProcessQueueLock struct {
	runningQueue     map[uint64]<-chan bool
	runningQueueLock *sync.RWMutex
	waitingQueue     map[uint64]chan bool
	waitingQueueLock *sync.RWMutex
	done             chan uint64
	wait             chan *LockRequest
}

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

// WaitingQueueWatceher - ...
func (p *ProcessQueueLock) WaitingQueueWatceher() {

	for {

		select {

		case num := <-p.done:

			p.waitingQueueLock.Lock()

			comm, ok := p.waitingQueue[num]
			if ok {

				comm <- true
				delete(p.waitingQueue, num)

			}

			p.waitingQueueLock.Unlock()

		case req := <-p.wait:

			p.waitingQueueLock.Lock()

			p.waitingQueue[req.Block] = req.Communication

			p.waitingQueueLock.Unlock()

		}

	}

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

		go p.Running(request)
		return true

	}

	p.wait <- request
	return false

}

// Running - ...
func (p *ProcessQueueLock) Running(request *LockRequest) {

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

// Release - When done with processing block, worker invokes method
// to let run tracker know it's done with processing this block, it there's
// any other worker waiting for processing same block, it can now go in
func (p *ProcessQueueLock) Release(comm chan<- bool) {

	comm <- true

}
