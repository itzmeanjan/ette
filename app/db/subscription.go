package db

import (
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/ethereum/go-ethereum/common"
	"gorm.io/gorm"
)

// DeliveryCountByPlanName - Given subscription plan name, returns subscription plan's delivery count
func DeliveryCountByPlanName(_db *gorm.DB, planName string) (uint64, error) {
	var subscriptionPlan SubscriptionPlans

	if err := _db.Where("name = ?", planName).Find(&subscriptionPlan).Error; err != nil {
		return 0, err
	}

	return subscriptionPlan.DeliveryCount, nil
}

// UpdateSubscriptionPlan - Tries to update existing subscription plan, where
// it's assumed plan name is unchanged & allowed delivery count in 24 hours
// has got updated
func UpdateSubscriptionPlan(_db *gorm.DB, name string, deliveryCount uint64) {

	if err := _db.Model(&SubscriptionPlans{}).Where("name = ?", name).Update("deliverycount", deliveryCount).Error; err != nil {
		log.Printf("[!] Failed to update subscription plan : %s\n", err.Error())
	}

}

// CreateSubscriptionPlan - Creates new entry for subscription plan
func CreateSubscriptionPlan(_db *gorm.DB, name string, deliveryCount uint64) {

	if err := _db.Create(&SubscriptionPlans{
		Name:          name,
		DeliveryCount: deliveryCount,
	}).Error; err != nil {
		log.Printf("[!] Failed to persist subscription plan : %s\n", err.Error())
	}

}

// AddNewSubscriptionPlan - Adding new subcription plan to database
// after those being read from .plans.json
//
// Taking into consideration the factor, whether it has
// been already persisted or not, or any changes made to `.plans.json` file
func AddNewSubscriptionPlan(_db *gorm.DB, name string, deliveryCount uint64) {

	count, err := DeliveryCountByPlanName(_db, name)
	// Plans not yet persisted in table, attempting to persist them 👇
	if err != nil {
		CreateSubscriptionPlan(_db, name, deliveryCount)
		return
	}

	switch count {
	case 0:
		// Entry doesn't yet exist, attempting to create it
		CreateSubscriptionPlan(_db, name, deliveryCount)
	case deliveryCount:
		// No change made in `.plans.json` file
		// i.e. subscription plan is already persisted
		return
	default:
		// Plan with same name already persisted in table
		// trying to update it
		UpdateSubscriptionPlan(_db, name, deliveryCount)
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

// GetAllowedDeliveryCountByAddress - Returns how many deliveries can be made to user
// in 24 hours, as per plan they're subscribed to
func GetAllowedDeliveryCountByAddress(_db *gorm.DB, address common.Address) uint64 {

	plan := CheckSubscriptionPlanDetailsByAddress(_db, address)
	if plan == nil {
		return 0
	}

	return plan.DeliveryCount

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
