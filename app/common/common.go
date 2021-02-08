package common

import "github.com/ethereum/go-ethereum/common"

// StringifyEventTopics - Given array of event topic signatures,
// returns their stringified form, to be required for publishing data to subscribers
// and persisting into database
func StringifyEventTopics(data []common.Hash) []string {
	buffer := make([]string, len(data))

	for i := 0; i < len(data); i++ {
		buffer[i] = data[i].Hex()
	}

	return buffer
}

// CreateEventTopicMap - Given array of event topics, returns map
// of valid event topics, to be used when performing selective field based
// queries on event topics
func CreateEventTopicMap(topics ...string) map[uint8]string {

	_topics := make(map[uint8]string)

	if topics[0] != "" {
		_topics[0] = topics[0]
	}

	if topics[1] != "" {
		_topics[1] = topics[1]
	}

	if topics[2] != "" {
		_topics[2] = topics[2]
	}

	if topics[3] != "" {
		_topics[3] = topics[3]
	}

	return _topics

}
