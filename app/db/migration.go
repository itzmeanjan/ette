package db

import (
	"log"

	"gorm.io/gorm"
)

// Migrate - Database model migrator
func Migrate(db *gorm.DB) {
	log.Fatalln("[!] ", db.AutoMigrate(&Blocks{}, &Transactions{}).Error())
}
