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

// ComputeAPIKeyForAddress - Computing new API key for user address, by taking nonce of that account under consideration
//
// Here nonce is nothing but count of applications created in `ette` by account
func ComputeAPIKeyForAddress(_db *gorm.DB, address common.Address) []byte {
	var count int64

	if err := _db.Model(&Users{}).Where("users.address = ?", address.Hex()).Count(&count).Error; err != nil {
		return nil
	}

	data, err := json.Marshal(&struct {
		Address common.Address `json:"address"`
		Nonce   int64          `json:"nonce"`
	}{
		Address: address,
		Nonce:   count + 1,
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
// to client application ( identified using signer address i.e. who created API Key ),
//  for query response or for real-time data delivery, is under a limit
// ( currently hardcoded inside code ) or not
func IsUnderRateLimit(_db *gorm.DB, userAddress string) bool {
	var count int64

	if err := _db.Model(&DeliveryHistory{}).
		Where("delivery_history.client = ? and delivery_history.ts > now() - interval '1 day'", userAddress).
		Count(&count).Error; err != nil {
		return false
	}

	// 50k times data delivered to client in a day
	return count < 50000
}
