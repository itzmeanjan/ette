package db

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/itzmeanjan/ette/app/data"
	"gorm.io/gorm"
)

// GetBlockByHash - Given blockhash finds out block related information
//
// If not found, returns nil
func GetBlockByHash(db *gorm.DB, hash common.Hash) *data.Block {
	var block data.Block

	if res := db.Model(&Blocks{}).Where("hash = ?", hash.Hex()).First(&block); res.Error != nil {
		return nil
	}

	return &block
}
