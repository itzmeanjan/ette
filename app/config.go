package app

import "github.com/spf13/viper"

// Reading .env file content, during application start up
func read(file string) error {
	viper.SetConfigFile(file)

	return viper.ReadInConfig()
}

// Get - Get config value by key
func Get(key string) string {
	return viper.GetString(key)
}
