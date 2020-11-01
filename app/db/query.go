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

// GetBlockByNumber - Fetch block using block number
//
// If not found, returns nil
func GetBlockByNumber(db *gorm.DB, number uint64) *data.Block {
	var block data.Block

	if res := db.Model(&Blocks{}).Where("number = ?", number).First(&block); res.Error != nil {
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
func GetBlocksByNumberRange(db *gorm.DB, from uint64, to uint64) *data.Blocks {
	var blocks []data.Block

	if res := db.Model(&Blocks{}).Where("number >= ? and number <= ?", from, to).Order("number asc").Find(&blocks); res.Error != nil {
		return nil
	}

	return &data.Blocks{
		Blocks: blocks,
	}
}

// GetBlocksByTimeRange - Given time range ( of 60 sec span at max ), returns blocks
// mined in that time span
//
// If asked to find out blocks in time span larger than 60 sec, simply drops query request
func GetBlocksByTimeRange(db *gorm.DB, from uint64, to uint64) *data.Blocks {
	var blocks []data.Block

	if res := db.Model(&Blocks{}).Where("time >= ? and time <= ?", from, to).Order("number asc").Find(&blocks); res.Error != nil {
		return nil
	}

	return &data.Blocks{
		Blocks: blocks,
	}
}

// GetTransactionsByBlockHash - Given block hash, returns all transactions
// present in that block
func GetTransactionsByBlockHash(db *gorm.DB, hash common.Hash) *data.Transactions {
	var tx []*data.Transaction

	if res := db.Model(&Transactions{}).Where("blockhash = ?", hash.Hex()).Find(&tx); res.Error != nil {
		return nil
	}

	return &data.Transactions{
		Transactions: tx,
	}
}

// GetTransactionsByBlockNumber - Given block number, returns all transactions
// present in that block
func GetTransactionsByBlockNumber(db *gorm.DB, number uint64) *data.Transactions {
	var tx []*data.Transaction

	if res := db.Model(&Transactions{}).Where("blockhash = (?)", db.Model(&Blocks{}).Where("number = ?", number).Select("hash")).Find(&tx); res.Error != nil {
		return nil
	}

	return &data.Transactions{
		Transactions: tx,
	}
}

// GetTransactionByHash - Given tx hash, extracts out transaction related data
func GetTransactionByHash(db *gorm.DB, hash common.Hash) *data.Transaction {
	var tx data.Transaction

	if err := db.Model(&Transactions{}).Where("hash = ?", hash.Hex()).First(&tx).Error; err != nil {
		return nil
	}

	return &tx
}

// GetTransactionsFromAccountByBlockNumberRange - Given account address & block number range, it can find out
// all transactions which are performed from this account
func GetTransactionsFromAccountByBlockNumberRange(db *gorm.DB, account common.Address, from uint64, to uint64) *data.Transactions {
	var tx []*data.Transaction

	if err := db.Model(&Transactions{}).Joins("left join blocks on transactions.blockhash = blocks.hash").Where("transactions.from = ? and blocks.number >= ? and blocks.number <= ?", account.Hex(), from, to).Select("transactions.hash, transactions.from, transactions.to, transactions.contract, transactions.gas, transactions.gasprice, transactions.cost, transactions.nonce, transactions.state, transactions.blockhash").Find(&tx).Error; err != nil {
		return nil
	}

	return &data.Transactions{
		Transactions: tx,
	}
}
