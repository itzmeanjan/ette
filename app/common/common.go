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
