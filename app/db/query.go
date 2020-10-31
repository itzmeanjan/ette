package db

import (
	"github.com/ethereum/go-ethereum/common"
	"gorm.io/gorm"
)

// GetBlockByHash - Given blockhash finds out block related information
//
// If not found, returns nil
func GetBlockByHash(db *gorm.DB, hash common.Hash) *Blocks {
	var block Blocks

	if res := db.Model(&Blocks{}).Where("hash = ?", hash.Hex()).First(&block); res.Error != nil {
		return nil
	}

	return &block
}
