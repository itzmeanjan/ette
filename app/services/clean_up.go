package services

import (
	"log"
	"time"

	"github.com/gookit/color"
	"github.com/itzmeanjan/ette/app/db"
	"gorm.io/gorm"
)

// DeliveryHistoryCleanUpService - This function is supposed to be run as
// an independent go routine, which will invoke self every 24 hours &
// attempt to clean up all delivery histories for all clients
// older than recent 24 hours
func DeliveryHistoryCleanUpService(_db *gorm.DB) {

	for {

		select {

		// This syntax invokes following function 24 hours later, last time it completed its job
		case <-time.After(time.Hour * time.Duration(24)):

			log.Printf(color.Green.Sprintf("[*] Attempting to clean delivery history older than 24 hours"))

			db.DropOldDeliveryHistories(_db)

			log.Printf(color.Green.Sprintf("[+] Cleaned delivery history older than 24 hours"))

		}

	}

}
