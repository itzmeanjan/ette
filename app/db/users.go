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

	if err := _db.Model(&Users{}).Where("users.address = ?", address.Hex()).Find(&apps).Error; err != nil {
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

	return true
}

// ValidateAPIKey - Given an API Key, checks whether this API Key is present
// or not, if yes, request from client can be taken up
func ValidateAPIKey(_db *gorm.DB, apiKey string) bool {
	var user Users

	if err := _db.Model(&Users{}).Where("users.apikey = ?", apiKey).First(&user).Error; err != nil {
		return false
	}

	return &user != nil
}

// IsUnderRateLimit - Checks whether number of times data delivered
// to client application ( identified using API Key ), for query response
// or for real-time data delivery, is under a limit ( currently hardcoded inside code )
// or not
func IsUnderRateLimit(_db *gorm.DB, apiKey string) bool {
	var count int64

	if err := _db.Model(&DeliveryHistory{}).
		Joins("inner join users on delivery_history.client = users.apikey").
		Where("users.address = (?)",
			_db.Model(&Users{}).Where("apikey = ?", apiKey).Select("address")).
		Where("delivery_history.ts > now() - interval '1 day'").Count(&count).Error; err != nil {
		return false
	}

	return count < 50000
}
