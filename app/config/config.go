package config

import (
	"log"
	"path/filepath"
	"strconv"

	"github.com/spf13/viper"
)

// Read - Reading .env file content, during application start up
func Read(file string) error {
	viper.SetConfigFile(file)

	return viper.ReadInConfig()
}

// Get - Get config value by key
func Get(key string) string {
	return viper.GetString(key)
}

// GetConcurrencyFactor - Reads concurrency factor specified in `.env` file, during deployment
// and returns that number as unsigned integer
func GetConcurrencyFactor() uint64 {

	factor := Get("ConcurrencyFactor")
	if factor == "" {
		return 1
	}

	parsedFactor, err := strconv.ParseUint(factor, 10, 64)
	if err != nil {
		log.Printf("[!] Failed to parse concurrency factor : %s\n", err.Error())
		return 1
	}

	return parsedFactor
}

// GetBlockConfirmations - Number of block confirmations required
// before considering that block to be finalized, and can be persisted
// in a permanent data store
func GetBlockConfirmations() uint64 {

	confirmationCount := Get("BlockConfirmations")
	if confirmationCount == "" {
		return 0
	}

	parsedConfirmationCount, err := strconv.ParseUint(confirmationCount, 10, 64)
	if err != nil {
		log.Printf("[!] Failed to parse block confirmations : %s\n", err.Error())
		return 0
	}

	return parsedConfirmationCount

}

// GetBlockNumberRange - Returns how many blocks can be queried at a time
// when performing range based queries from client side
func GetBlockNumberRange() uint64 {

	blockRange := Get("BlockRange")
	if blockRange == "" {
		return 100
	}

	parsedBlockRange, err := strconv.ParseUint(blockRange, 10, 64)
	if err != nil {
		log.Printf("[!] Failed to parse block range : %s\n", err.Error())
		return 100
	}

	return parsedBlockRange

}

// GetTimeRange - Returns what's the max time span that can be used while performing query
// from client side, in terms of second
func GetTimeRange() uint64 {

	timeRange := Get("TimeRange")
	if timeRange == "" {
		return 3600
	}

	parsedTimeRange, err := strconv.ParseUint(timeRange, 10, 64)
	if err != nil {
		log.Printf("[!] Failed to parse time range : %s\n", err.Error())
		return 3600
	}

	return parsedTimeRange

}

// GetSnapshotFile - Reading snapshot file name from
// config file, if not provided, `snapshot.bin` is used as default file name
func GetSnapshotFile() string {

	_file := Get("SnapshotFile")
	if _file == "" {
		_file = "snapshot.bin"
	}

	_absFile, err := filepath.Abs(_file)
	if err != nil {
		log.Fatalf("[!] Failed to find real path of `%s` : %s\n", _file, err.Error())
	}

	return _absFile

}
