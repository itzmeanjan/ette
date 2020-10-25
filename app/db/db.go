package db

import (
	"fmt"
	"log"
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	cfg "github.com/itzmeanjan/ette/app/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Connect - Connecting to postgresql database
func Connect() *gorm.DB {
	port, err := strconv.Atoi(cfg.Get("DB_PORT"))
	if err != nil {
		log.Fatalln("[!] ", err)
	}

	_db, err := gorm.Open(postgres.Open(fmt.Sprintf("postgresql://%s:%s@%s:%d/%s", cfg.Get("DB_USER"), cfg.Get("DB_PASSWORD"), cfg.Get("DB_HOST"), port, cfg.Get("DB_NAME"))), &gorm.Config{})
	if err != nil {
		log.Fatalln("[!] ", err)
	}

	_db.AutoMigrate(&Blocks{}, &Transactions{})
	return _db
}

// GetBlock - Fetch block by number, from database
func GetBlock(_db *gorm.DB, number *big.Int) *Blocks {
	var block Blocks

	if err := _db.Where("number = ?", number.String()).First(&block).Error; err != nil {
		log.Println("[!] ", err)
		return nil
	}

	return &block
}

// PutBlock - Persisting fetched block information in database
func PutBlock(_db *gorm.DB, _block *types.Block) {
	if err := _db.Create(&Blocks{
		Hash:       _block.Hash().Hex(),
		Number:     _block.Number().String(),
		Time:       _block.Time(),
		ParentHash: _block.ParentHash().Hex(),
		Difficulty: _block.Difficulty().String(),
		GasUsed:    _block.GasUsed(),
		GasLimit:   _block.GasLimit(),
		Nonce:      _block.Nonce(),
	}).Error; err != nil {
		log.Println("[!] ", err)
	}
}

// GetTransaction - Fetches tx entry from database, given txhash & containing block hash
func GetTransaction(_db *gorm.DB, blkHash common.Hash, txHash common.Hash) *Transactions {
	var tx Transactions

	if err := _db.Where("hash = ? and blockhash = ?", txHash.Hex(), blkHash.Hex()).First(&tx).Error; err != nil {
		log.Println("[!] ", err)
		return nil
	}

	return &tx
}

// PutTransaction - Persisting transactions present in a block in database
func PutTransaction(_db *gorm.DB, _tx *types.Transaction, _txReceipt *types.Receipt, _sender common.Address) {
	if err := _db.Create(&Transactions{
		Hash:      _tx.Hash().Hex(),
		From:      _sender.Hex(),
		To:        _tx.To().Hex(),
		Gas:       _tx.Gas(),
		GasPrice:  _tx.GasPrice().String(),
		Cost:      _tx.Cost().String(),
		Nonce:     _tx.Nonce(),
		State:     _txReceipt.Status,
		BlockHash: _txReceipt.BlockHash.Hex(),
	}); err != nil {
		log.Println("[!] ", err)
	}
}
