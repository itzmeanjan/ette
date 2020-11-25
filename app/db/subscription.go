package db

import (
	"log"

	"gorm.io/gorm"
)

// AddNewSubscriptionPlan - Adding new subcription plan to database
// after those being read from .plans.json
func AddNewSubscriptionPlan(_db *gorm.DB, name string, deliveryCount uint64) {

	if err := _db.Create(&SubscriptionPlans{
		Name:          name,
		DeliveryCount: deliveryCount,
	}).Error; err != nil {
		log.Printf("[!] Failed to add subscription plan : %s\n", err.Error())
	}

}
