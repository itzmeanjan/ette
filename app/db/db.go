package db

import (
	"fmt"
	"log"
	"strconv"

	cfg "github.com/itzmeanjan/ette/app/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Connect - Connecting to postgresql database
func Connect() *gorm.DB {
	port, err := strconv.Atoi(cfg.Get("DB_PORT"))
	if err != nil {
		log.Fatalln("[!] ", err)
	}

	db, err := gorm.Open(postgres.Open(fmt.Sprintf("postgresql://%s:%s@%s:%d/%s", cfg.Get("DB_USER"), cfg.Get("DB_PASSWORD"), cfg.Get("DB_HOST"), port, cfg.Get("DB_NAME"))), &gorm.Config{})
	if err != nil {
		log.Fatalln("[!] ", err)
	}

	return db
}
