package db

import (
	"fmt"
	"log"

	cfg "github.com/itzmeanjan/ette/app/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connect - Connecting to postgresql database
func Connect() *gorm.DB {
	_db, err := gorm.Open(postgres.Open(fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", cfg.Get("DB_USER"), cfg.Get("DB_PASSWORD"), cfg.Get("DB_HOST"), cfg.Get("DB_PORT"), cfg.Get("DB_NAME"))), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		log.Fatalf("[!] Failed to connect to db : %s\n", err.Error())
	}

	_db.AutoMigrate(&Blocks{}, &Transactions{}, &Events{}, &DeliveryHistory{}, &Users{})
	return _db
}
