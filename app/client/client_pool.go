package client

import (
	"errors"
	"sync"
	"time"
)

// Pool Creates many HTTP clients and WebSocket clients to rotate RPC clients when
// one of them has an error
//
// We set this client as unavailable
type Pool struct {
	httpClients      []*Client
	webSocketClients []*Client
	mu               sync.Mutex
	retryTimes       int
}

// DefaultClientPool Create the default pool with specific options
func DefaultClientPool() *Pool {
	httpClients, webSocketClients := getClients()
	clientPool := NewClientPool(
		WithHttpClients(httpClients),
		WithWebsocketClients(webSocketClients),
		WithRetryTimes(5),
	)
	return clientPool
}

// NewClientPool Create pool with many options
func NewClientPool(options ...func(*Pool)) *Pool {
	pool := DefaultClientPool()

	for _, o := range options {
		o(pool)
	}
	return pool
}

// WithHttpClients create pool with http clients
func WithHttpClients(httpClients []*Client) func(*Pool) {
	return func(pool *Pool) {
		pool.httpClients = httpClients
	}
}

// WithWebsocketClients create pool with websocket clients
func WithWebsocketClients(websocketClients []*Client) func(*Pool) {
	return func(pool *Pool) {
		pool.webSocketClients = websocketClients
	}
}

// WithRetryTimes create pool with retry times to get the available client
func WithRetryTimes(retryTimes int) func(*Pool) {
	return func(pool *Pool) {
		pool.retryTimes = retryTimes
	}
}

// GetAvailableClient get the currently available client
// If we try to get the available client many times, we return an error to check the RPC configuration
// If the client is available, we use this to interact with the blockchain node
func (p *Pool) GetAvailableClient(isRPC bool) (*Client, error) {
	var clients []*Client
	if isRPC {
		clients = p.httpClients
	} else {
		clients = p.webSocketClients
	}
	// Lock the process to prevent concurrently picking some unavailable clients
	p.mu.Lock()
	defer p.mu.Unlock()
	retried := 0
	for retried <= p.retryTimes {
		for _, c := range clients {
			if c.isAvailable() {
				return c, nil
			}
		}
		// If there is no client available at this time, sleep 1s to retry again
		time.Sleep(1 * time.Second)
		retried++
	}
	return nil, errors.New("[!] there are no clients available, please check the RPCs configuration")
}
