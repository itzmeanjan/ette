package common

import (
	"errors"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
)

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
func CreateEventTopicMap(topics []string) map[uint8]string {

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

// ParseNumber - Given an integer as string, attempts to parse it
func ParseNumber(number string) (uint64, error) {

	_num, err := strconv.ParseUint(number, 10, 64)
	if err != nil {

		return 0, errors.New("Failed to parse integer")

	}

	return _num, nil

}

// RangeChecker - Checks whether given number range is at max
// `limit` far away
func RangeChecker(from string, to string, limit uint64) (uint64, uint64, error) {

	_from, err := ParseNumber(from)
	if err != nil {
		return 0, 0, errors.New("Failed to parse integer")
	}

	_to, err := ParseNumber(to)
	if err != nil {
		return 0, 0, errors.New("Failed to parse integer")
	}

	if !(_to-_from < limit) {
		return 0, 0, errors.New("Range too long")
	}

	return _from, _to, nil

}
