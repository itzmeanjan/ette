package block

import "sync"

// ProcessQueueLock - ...
type ProcessQueueLock struct {
	RunningQueue     map[uint64]chan bool
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
func (p *ProcessQueueLock) Running(block uint64, comm chan bool) {

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

	<-comm
	return

}

// Release - ...
func (p *ProcessQueueLock) Release() {}
