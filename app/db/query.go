package db

import (
	"strconv"

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

// GetBlockByNumber - Fetch block using block number
//
// If not found, returns nil
func GetBlockByNumber(db *gorm.DB, number string) *data.Block {
	_num, err := strconv.Atoi(number)
	if err != nil {
		return nil
	}

	var block data.Block

	if res := db.Model(&Blocks{}).Where("number = ?", _num).First(&block); res.Error != nil {
		return nil
	}

	return &block
}
