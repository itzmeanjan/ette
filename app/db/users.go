package db

import (
	"github.com/ethereum/go-ethereum/common"
	"gorm.io/gorm"
)

// GetAppsByUserAddress - Given user address, returns list of all
// apps created by user, along with their API key & respective creation timestamp
func GetAppsByUserAddress(_db *gorm.DB, address common.Address) []*Users {
	var apps []*Users

	if err := _db.Model(&Users{}).Where("users.address = ?", address.Hex()).Find(&apps).Error; err != nil {
		return nil
	}

	return apps
}
