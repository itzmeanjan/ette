package snapshot

import (
	"fmt"
	"strings"

	"github.com/itzmeanjan/ette/app/data"
	"github.com/itzmeanjan/ette/app/db"
	pb "github.com/itzmeanjan/ette/app/pb"
)

// EventToProtoBuf - Creating proto buffer compatible data
// format for event data, which can be easily serialized & deserialized
// for taking snapshot and restoring from it
func EventToProtoBuf(event *data.Event) *pb.Event {

	return &pb.Event{
		BlockHash:       event.BlockHash,
		Index:           uint32(event.Index),
		Origin:          event.Origin,
		Topics:          strings.Fields(fmt.Sprintf("%q", event.Topics)),
		Data:            event.Data,
		TransactionHash: event.TransactionHash,
	}

}

// EventsToProtoBuf - Creating proto buffer compatible data
// format for events data, which can be easily serialized & deserialized
// for taking snapshot and restoring from it
func EventsToProtoBuf(events *data.Events) []*pb.Event {

	_events := make([]*pb.Event, len(events.Events))

	for i := 0; i < len(events.Events); i++ {
		_events[i] = EventToProtoBuf(events.Events[i])
	}

	return _events

}

// ProtoBufToEvent - This function will be helpful in constructing
// struct, which can be used for recovering from snapshotted data
func ProtoBufToEvent(event *pb.Event) *db.Events {

	return &db.Events{
		BlockHash:       event.BlockHash,
		Index:           uint(event.Index),
		Origin:          event.Origin,
		Topics:          event.Topics,
		Data:            event.Data,
		TransactionHash: event.TransactionHash,
	}

}

// ProtoBufToEvents - Will be helpful while deserializing from snapshot
// and attempting to restore into DB
func ProtoBufToEvents(events []*pb.Event) []*db.Events {

	_events := make([]*db.Events, len(events))

	for k, v := range events {

		_events[k] = ProtoBufToEvent(v)

	}

	return _events

}
