package db

import (
	"encoding/json"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"gorm.io/gorm"
)

// GetAppsByUserAddress - Given user address, returns list of all
// apps created by user, along with their API key & respective creation timestamp
func GetAppsByUserAddress(_db *gorm.DB, address common.Address) []*Users {
	var apps []*Users

	if err := _db.Model(&Users{}).Where("users.address = ?", address.Hex()).Order("users.ts desc").Find(&apps).Error; err != nil {
		return nil
	}

	if len(apps) == 0 {
		return nil
	}

	return apps
}

// ComputeAPIKeyForAddress - Computing new API key for user address,
// by taking `nonce` of that account & current unix time stamp ( with nanosecond level precision )
//  under consideration
//
// Here `nonce` is nothing but count of applications created in `ette` by account
//
// `Time` field is being added when computing next API key for making it more
// unpredictable
//
// Previous implementation had very predicatable pattern
func ComputeAPIKeyForAddress(_db *gorm.DB, address common.Address) []byte {
	var count int64

	if err := _db.Model(&Users{}).Where("users.address = ?", address.Hex()).Count(&count).Error; err != nil {
		return nil
	}

	data, err := json.Marshal(&struct {
		Address common.Address `json:"address"`
		Nonce   int64          `json:"nonce"`
		Time    int64          `json:"time"`
	}{
		Address: address,
		Nonce:   count + 1,
		Time:    time.Now().UTC().UnixNano(),
	})
	if err != nil {
		return nil
	}

	return crypto.Keccak256(data)
}

// RegisterNewApp - Registering new application for given address
func RegisterNewApp(_db *gorm.DB, address common.Address) bool {
	apiKey := ComputeAPIKeyForAddress(_db, address)
	if apiKey == nil {
		return false
	}

	if err := _db.Create(&Users{
		Address:   address.Hex(),
		APIKey:    common.BytesToHash(apiKey).Hex(),
		TimeStamp: time.Now().UTC(),
	}).Error; err != nil {
		return false
	}

	if CheckSubscriptionPlanByAddress(_db, address) != nil {
		return true
	}

	planID := GetDefaultSubscriptionPlanID(_db)
	if planID == 0 {
		return false
	}

	return AddSubscriptionPlanForAddress(_db, address, planID)
}

// ToggleAPIKeyState - Given valid API key, toggles its enabled state
func ToggleAPIKeyState(_db *gorm.DB, apiKey string) bool {
	user := GetUserFromAPIKey(_db, apiKey)
	if user == nil {
		return false
	}

	if err := _db.Model(&Users{}).Where("users.address = ? and users.apikey = ?", user.Address, user.APIKey).Update("enabled", !user.Enabled).Error; err != nil {
		return false
	}

	return true
}

// GetUserFromAPIKey - Given API Key, tries to find out if there's any user registered
// who signed for creating this API Key
func GetUserFromAPIKey(_db *gorm.DB, apiKey string) *Users {
	var user Users

	if err := _db.Model(&Users{}).Where("users.apikey = ?", apiKey).First(&user).Error; err != nil {
		return nil
	}

	return &user
}

// ValidateAPIKey - Given an API Key, checks whether this API Key is present
// or not, if yes, request from client can be taken up
func ValidateAPIKey(_db *gorm.DB, apiKey string) bool {
	return GetUserFromAPIKey(_db, apiKey) != nil
}

// IsUnderRateLimit - Checks whether number of times data delivered
// to client application on this day ( identified using signer address
// i.e. who created API Key ), for query response or for real-time data delivery,
// is under a limit ( currently hardcoded inside code ) or not
func IsUnderRateLimit(_db *gorm.DB, userAddress string) bool {

	// Getting current local time of machine
	now := time.Now()

	// Segregating into parts required
	var (
		day   = now.Day()
		month = now.Month()
		year  = now.Year()
	)

	var count int64

	if err := _db.Model(&DeliveryHistory{}).
		Where("delivery_history.client = ? and extract(day from delivery_history.ts) = ? and extract(month from delivery_history.ts) = ? and extract(year from delivery_history.ts) = ?", userAddress, day, month, year).
		Count(&count).Error; err != nil {
		return false
	}

	// Compare it with allowed rate count per 24 hours, of plan user is subscribed to
	return count < int64(GetAllowedDeliveryCountByAddress(_db, common.HexToAddress(userAddress)))
}

// DropOldDeliveryHistories - Attempts to delete older than 24 hours delivery history
// from data store, because that piece of data is not being used any where, so no need to keep it
//
// Before delivering any piece of data to client, rate limit to be checked for last 24 hours
// not older than that
//
// @note This function can be invoked every 24 hours to clean up historical entries
func DropOldDeliveryHistories(_db *gorm.DB) {

	// Wrapping delete operation inside DB transaction to make it
	// consistent to other parties attempting to read from same table
	_db.Transaction(func(dbWtx *gorm.DB) error {

		return dbWtx.Where("delivery_history.ts < now() - interval '24 hours'").Delete(&DeliveryHistory{}).Error

	})

}
