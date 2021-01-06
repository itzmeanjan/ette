package db

import (
	"log"

	"github.com/ethereum/go-ethereum/core/types"
	d "github.com/itzmeanjan/ette/app/data"
	"gorm.io/gorm"
)

// StoreBlock - Persisting block data in database,
// if data already not stored
//
// Also checks equality with existing data, if mismatch found
// updated with latest data
func StoreBlock(_db *gorm.DB, _block *types.Block, _status *d.StatusHolder) bool {

	persistedBlock := GetBlock(_db, _block.NumberU64())
	if persistedBlock == nil {

		// If we're able to successfully insert block data into table
		// it's going to be considered, while calculating total block count in database
		status := PutBlock(_db, _block)
		if status {
			// Trying to safely update inserted block count
			// -- Critical section of code
			_status.IncrementBlocksInserted()
			// -- ends here
		}

		return status

	}

	if !persistedBlock.SimilarTo(_block) {
		return UpdateBlock(_db, _block)
	}

	return true
}

// GetBlock - Fetch block by number, from database
func GetBlock(_db *gorm.DB, number uint64) *Blocks {
	var block Blocks

	if err := _db.Where("number = ?", number).First(&block).Error; err != nil {
		return nil
	}

	return &block
}

// PutBlock - Persisting fetched block information in database
func PutBlock(_db *gorm.DB, _block *types.Block) bool {
	status := true

	if err := _db.Create(&Blocks{
		Hash:                _block.Hash().Hex(),
		Number:              _block.NumberU64(),
		Time:                _block.Time(),
		ParentHash:          _block.ParentHash().Hex(),
		Difficulty:          _block.Difficulty().String(),
		GasUsed:             _block.GasUsed(),
		GasLimit:            _block.GasLimit(),
		Nonce:               _block.Nonce(),
		Miner:               _block.Coinbase().Hex(),
		Size:                float64(_block.Size()),
		TransactionRootHash: _block.TxHash().Hex(),
		ReceiptRootHash:     _block.ReceiptHash().Hex(),
	}).Error; err != nil {
		status = false
		log.Printf("[!] Failed to persist block : %d : %s\n", _block.NumberU64(), err.Error())
	}

	return status
}

// UpdateBlock - Updating already existing block entry with newly
// obtained info
func UpdateBlock(_db *gorm.DB, _block *types.Block) bool {
	status := true

	if err := _db.Where("number = ?", _block.NumberU64()).Updates(&Blocks{
		Hash:                _block.Hash().Hex(),
		Time:                _block.Time(),
		ParentHash:          _block.ParentHash().Hex(),
		Difficulty:          _block.Difficulty().String(),
		GasUsed:             _block.GasUsed(),
		GasLimit:            _block.GasLimit(),
		Nonce:               _block.Nonce(),
		Miner:               _block.Coinbase().Hex(),
		Size:                float64(_block.Size()),
		TransactionRootHash: _block.TxHash().Hex(),
		ReceiptRootHash:     _block.ReceiptHash().Hex(),
	}).Error; err != nil {
		status = false
		log.Printf("[!] Failed to update block : %d : %s\n", _block.NumberU64(), err.Error())
	}

	return status
}
