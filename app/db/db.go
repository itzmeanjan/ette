package db

import (
	"fmt"
	"log"
	"strconv"

	"github.com/itzmeanjan/ette/app"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Connect - Connecting to postgresql database
func Connect() *gorm.DB {
	port, err := strconv.Atoi(app.Get("DB_PORT"))
	if err != nil {
		log.Fatalln("[!] ", err)
	}

	db, err := gorm.Open(postgres.Open(fmt.Sprintf("postgresql://%s:%s@%s:%d/%s", app.Get("DB_USER"), app.Get("DB_PASSWORD"), app.Get("DB_HOST"), port, app.Get("DB_NAME"))), &gorm.Config{})
	if err != nil {
		log.Fatalln("[!] ", err)
	}

	return db
}
