package config

import (
	"log"
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

	if !(parsedFactor > 0) {
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

	if !(parsedConfirmationCount > 0) {
		return 0
	}

	return parsedConfirmationCount

}
