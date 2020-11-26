package db

import (
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/ethereum/go-ethereum/common"
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

// PersistAllSubscriptionPlans - Given path to user created subscription plan
// holder ``.plans.json` file, it'll read that content into memory & then parse JSON
// content of its, which will be persisted into database, into `subscription_plans` table
func PersistAllSubscriptionPlans(_db *gorm.DB, file string) {

	data, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatalf("[!] Failed to read content from subscription plan file : %s\n", err.Error())
	}

	type Plan struct {
		Name          string `json:"name"`
		DeliveryCount uint64 `json:"deliveryCount"`
	}

	type Plans struct {
		Plans []*Plan `json:"plans"`
	}

	var plans Plans

	if err := json.Unmarshal(data, &plans); err != nil {
		log.Fatalf("[!] Failed to parse JSON content from subscription plan file : %s\n", err.Error())
	}

	for _, v := range plans.Plans {
		AddNewSubscriptionPlan(_db, v.Name, v.DeliveryCount)
	}

	log.Printf("[+] Successfully persisted subscription plans into database")

}

// GetAllSubscriptionPlans - Returns a list of all available susbcription plans
// from this `ette` instance
func GetAllSubscriptionPlans(_db *gorm.DB) []*SubscriptionPlans {
	var plans []*SubscriptionPlans

	if err := _db.Model(&SubscriptionPlans{}).Find(&plans).Error; err != nil {
		return nil
	}

	return plans
}

// CheckSubscriptionPlanByAddress - Given user's ethereum address, checking if user has
// already subscribed to any plan or not
func CheckSubscriptionPlanByAddress(_db *gorm.DB, address common.Address) *SubscriptionDetails {
	var details SubscriptionDetails

	if err := _db.Model(&SubscriptionDetails{}).Where("address = ?", address.Hex()).First(&details).Error; err != nil {
		return nil
	}

	return &details
}
