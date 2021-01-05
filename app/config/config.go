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

	return parsedFactor
}
