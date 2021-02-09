package block

import "sync"

// ProcessQueueLock - ...
type ProcessQueueLock struct {
	runningQueue     map[uint64]<-chan bool
	runningQueueLock *sync.RWMutex
	waitingQueue     map[uint64]chan bool
	waitingQueueLock *sync.RWMutex
	done             <-chan uint64
	wait             <-chan bool
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

		case <-p.done:

		case <-p.wait:

		}

	}

}

// Acquire - ...
func (p *ProcessQueueLock) Acquire(block uint64, comm chan bool) bool {

	_, ok := p.runningQueue[block]
	if !ok {

		// -- Accessing critical section of code, acquiring lock
		p.runningQueueLock.Lock()

		p.runningQueue[block] = comm

		p.runningQueueLock.Unlock()
		// -- Released lock, done with critical section of code

		go p.Running(block, comm)
		return true

	}
	// -- Accessing critical section of code, acquiring lock
	p.waitingQueueLock.Lock()

	p.waitingQueue[block] = comm

	p.waitingQueueLock.Unlock()
	// -- Released lock, done with critical section of code

	return false

}

// Running - ...
func (p *ProcessQueueLock) Running(block uint64, comm <-chan bool) {

	defer func() {

		// -- Critical section of code, below
		p.runningQueueLock.Lock()

		delete(p.runningQueue, block)

		p.runningQueueLock.Unlock()
		// -- Releasing lock, done with critical section of code

	}()

	// this is a blocking call, only to be unblocked when
	// client has done it's job & call release lock
	<-comm
	return

}

// Release - When done with processing block, worker invokes method
// to let run tracker know it's done with processing this block, it there's
// any other worker waiting for processing same block, it can now go in
func (p *ProcessQueueLock) Release(comm chan<- bool) {

	comm <- true

}
