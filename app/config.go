package app

import "github.com/spf13/viper"

// Reading .env file content, during application start up
func read(file string) error {
	viper.SetConfigFile(file)

	return viper.ReadInConfig()
}

// Get config value by key
func get(key string) string {
	return viper.GetString(key)
}
