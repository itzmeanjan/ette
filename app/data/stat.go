package data

import "sync/atomic"

// ActiveSubscriptions - Keeps track of how many active websocket
// connections being maintained now by `ette`
type ActiveSubscriptions struct {
	Count uint64
}

// Increment - Safely increment count by `X`
func (a *ActiveSubscriptions) Increment(by uint64) {
	atomic.AddUint64(&a.Count, by)
}

// Decrement - Safely decrement count by `X`
func (a *ActiveSubscriptions) Decrement(by uint64) {
	atomic.AddUint64(&a.Count, ^uint64(by-1))
}

// SendReceiveCounter - Keeps track of how many read & write ops
// were performed to & from socket during life time of one single
// websocket connection
type SendReceiveCounter struct {
	Send    uint64
	Receive uint64
}

// IncrementSend -To be invoked when new data written into socket
func (s *SendReceiveCounter) IncrementSend(by uint64) {
	atomic.AddUint64(&s.Send, by)
}

// IncrementReceive - To be invoked when new data read from socket
func (s *SendReceiveCounter) IncrementReceive(by uint64) {
	atomic.AddUint64(&s.Receive, by)
}
