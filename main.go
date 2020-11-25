package main

import (
	"log"
	"path/filepath"

	"github.com/itzmeanjan/ette/app"
)

func main() {
	configFile, err := filepath.Abs(".env")
	if err != nil {
		log.Fatalf("[!] Failed to find `.env` : %s\n", err.Error())
	}

	subscriptionPlansFile, err := filepath.Abs(".plans.json")
	if err != nil {
		log.Fatalf("[!] Failed to find `.plans.json` : %s\n", err.Error())
	}

	app.Run(configFile, subscriptionPlansFile)
}
