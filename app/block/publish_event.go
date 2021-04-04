package block

import (
	"context"
	"log"

	d "github.com/itzmeanjan/ette/app/data"
	"github.com/itzmeanjan/ette/app/db"
)

// PublishEvents - Iterate over all events & try to publish them on
// redis pubsub channel
func PublishEvents(blockNumber uint64, events []*db.Events, redis *d.RedisInfo) bool {

	if events == nil {
		return false
	}

	var status bool

	for _, e := range events {

		status = PublishEvent(blockNumber, e, redis)
		if !status {
			break
		}

	}

	return status

}

// PublishEvent - Publishing event/ log entry to redis pub-sub topic, to be captured by subscribers
// and sent to client application, who are interested in this piece of data
// after applying filter
func PublishEvent(blockNumber uint64, event *db.Events, redis *d.RedisInfo) bool {

	if event == nil {
		return false
	}

	data := &d.Event{
		Origin:          event.Origin,
		Index:           event.Index,
		Topics:          event.Topics,
		Data:            event.Data,
		TransactionHash: event.TransactionHash,
		BlockHash:       event.BlockHash,
	}

	if err := redis.Client.Publish(context.Background(), redis.EventPublishTopic, data).Err(); err != nil {

		log.Printf("❗️ Failed to publish event from block %d : %s\n", blockNumber, err.Error())
		return false

	}

	return true

}
