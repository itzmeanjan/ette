package data

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
)

// EventConsumer - Event consumption to be managed by this struct, when new websocket
// connection requests for receiving event data, it'll create this struct, with necessary pieces
// of information, which is to be required when delivering data & checking whether this connection
// has really requested notification for this event or not
type EventConsumer struct {
	Client     *redis.Client
	Request    *SubscriptionRequest
	Connection *websocket.Conn
	PubSub     *redis.PubSub
}

// Subscribe - Event consumer is subscribing to `event` topic,
// where all event related data to be published
func (e *EventConsumer) Subscribe() {
	e.PubSub = e.Client.Subscribe(context.Background(), e.Request.Topic())
}
