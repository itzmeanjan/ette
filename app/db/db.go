package db

import (
	"fmt"
	"log"
	"time"

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

// PutDataDeliveryInfo - Persisting data delivery info, before it's sent to client application
//
// dataLength is length of data in bytes, sent to client application
func PutDataDeliveryInfo(_db *gorm.DB, client string, endPoint string, dataLength uint64) {
	if err := _db.Create(&DeliveryHistory{
		Client:     client,
		TimeStamp:  time.Now().UTC(),
		EndPoint:   endPoint,
		DataLength: dataLength,
	}).Error; err != nil {
		log.Printf("[!] Failed to persist data delivery info : %s\n", err.Error())
	}
}
