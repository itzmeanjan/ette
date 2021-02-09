package block

import "sync"

// ProcessQueueLock - ...
type ProcessQueueLock struct {
	RunningQueue     map[uint64]<-chan bool
	RunningQueueLock *sync.RWMutex
	WaitingQueue     map[uint64]chan bool
	WaitingQueueLock *sync.RWMutex
}

// Acquire - ...
func (p *ProcessQueueLock) Acquire(block uint64, comm chan bool) {

	_, ok := p.RunningQueue[block]
	if !ok {

	}

}

// Running - ...
func (p *ProcessQueueLock) Running(block uint64, comm <-chan bool) {

	defer func() {

		// -- Critical section of code, below
		p.RunningQueueLock.Lock()

		delete(p.RunningQueue, block)

		p.RunningQueueLock.Unlock()
		// -- Releasing lock, done with critical section of code

	}()

	// -- Accessing critical section of code, acquiring lock
	p.RunningQueueLock.Lock()

	p.RunningQueue[block] = comm

	p.RunningQueueLock.Unlock()
	// -- Released lock, done with critical section of code

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
