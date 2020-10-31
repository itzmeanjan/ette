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
	_num, err := strconv.ParseUint(number, 10, 64)
	if err != nil {
		return nil
	}

	var block data.Block

	if res := db.Model(&Blocks{}).Where("number = ?", _num).First(&block); res.Error != nil {
		return nil
	}

	return &block
}

// GetBlocksByNumberRange - Given block numbers as range, it'll extract out those blocks
// by number, while returning them in ascendically sorted form in terms of block numbers
//
// Note : Can return at max 10 blocks in a single query
//
// If more blocks are requested, simply to be rejected
// In that case, consider splitting them such that they satisfy criteria
func GetBlocksByNumberRange(db *gorm.DB, from string, to string) *data.Blocks {
	_fromNum, err := strconv.ParseUint(from, 10, 64)
	if err != nil {
		return nil
	}

	_toNum, err := strconv.ParseUint(to, 10, 64)
	if err != nil {
		return nil
	}

	if !(_toNum-_fromNum < 10) {
		return nil
	}

	var blocks data.Blocks

	if res := db.Model(&Blocks{}).Where("number >= ? and number <= ?", _fromNum, _toNum).Order("number asc").Find(&blocks); res.Error != nil {
		return nil
	}

	return &blocks
}

// GetBlocksByTimeRange - Given time range ( of 60 sec span at max ), returns blocks
// mined in that time span
//
// If asked to find out blocks in time span larger than 60 sec, simply drops query request
func GetBlocksByTimeRange(db *gorm.DB, from string, to string) *data.Blocks {
	_fromTime, err := strconv.ParseUint(from, 10, 64)
	if err != nil {
		return nil
	}

	_toTime, err := strconv.ParseUint(to, 10, 64)
	if err != nil {
		return nil
	}

	if !(_toTime-_fromTime < 60) {
		return nil
	}

	var blocks data.Blocks

	if res := db.Model(&Blocks{}).Where("time >= ? and time <= ?", _fromTime, _toTime).Order("number asc").Find(&blocks); res.Error != nil {
		return nil
	}

	return &blocks
}
