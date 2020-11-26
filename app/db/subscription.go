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

// CheckSubscriptionPlanDetailsByAddress - Given address of subscriber, returns full plan details
// to which they're subscribed
func CheckSubscriptionPlanDetailsByAddress(_db *gorm.DB, address common.Address) *SubscriptionPlans {
	details := CheckSubscriptionPlanByAddress(_db, address)
	if details == nil {
		return nil
	}

	var plan SubscriptionPlans

	if err := _db.Model(&SubscriptionPlans{}).Where("id = ?", details.SubscriptionPlan).First(&plan).Error; err != nil {
		return nil
	}

	return &plan
}

// IsValidSubscriptionPlan - Given subscription plan id, checking against
// database whether it's a valid one or not
func IsValidSubscriptionPlan(_db *gorm.DB, id uint32) bool {
	var plan SubscriptionPlans

	if err := _db.Model(&SubscriptionPlans{}).Where("id = ?", id).First(&plan).Error; err != nil {
		return false
	}

	return plan.ID == id
}

// GetDefaultSubscriptionPlanID - Finding out that subscription plan id, which has lowest daily deliveryCount
// promise, which is going to be always default plan, when a new ethereum address joins `ette`
func GetDefaultSubscriptionPlanID(_db *gorm.DB) uint32 {
	var plan SubscriptionPlans

	if err := _db.Model(&SubscriptionPlans{}).Where("deliverycount = (?)", _db.Model(&SubscriptionPlans{}).Select("min(deliverycount)")).Select("id").First(&plan).Error; err != nil {
		return 0
	}

	return plan.ID
}

// AddSubscriptionPlanForAddress - Persisting subscription plan for one ethereum address, when this address
// is first time creating one `ette` application
func AddSubscriptionPlanForAddress(_db *gorm.DB, address common.Address, planID uint32) bool {
	if err := _db.Create(&SubscriptionDetails{
		Address:          address.Hex(),
		SubscriptionPlan: planID,
	}).Error; err != nil {
		return false
	}

	return true
}
