package main

import (
	"log"
	"path/filepath"

	"github.com/itzmeanjan/ette/app"
)

func main() {
	abs, err := filepath.Abs(".env")
	if err != nil {
		log.Fatalln("[!] ", err)
	}

	app.Run(abs)
}
