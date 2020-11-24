package db

import (
	"log"

	"github.com/ethereum/go-ethereum/core/types"
	"gorm.io/gorm"
)

// GetBlock - Fetch block by number, from database
func GetBlock(_db *gorm.DB, number uint64) *Blocks {
	var block Blocks

	if err := _db.Where("number = ?", number).First(&block).Error; err != nil {
		return nil
	}

	return &block
}

// PutBlock - Persisting fetched block information in database
func PutBlock(_db *gorm.DB, _block *types.Block) {
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
		log.Printf("[!] Failed to persist block : %d : %s\n", _block.NumberU64(), err.Error())
	}
}
