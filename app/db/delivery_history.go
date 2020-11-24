package db

import (
	"log"
	"time"

	"gorm.io/gorm"
)

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
