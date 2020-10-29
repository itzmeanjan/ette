package main

import (
	"log"
	"path/filepath"

	"github.com/itzmeanjan/ette/app"
)

func main() {
	abs, err := filepath.Abs(".env")
	if err != nil {
		log.Fatalf("[!] Failed to find `.env` : %s\n", err.Error())
	}

	app.Run(abs)
}
