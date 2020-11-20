package db

import (
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	com "github.com/itzmeanjan/ette/app/common"
	cfg "github.com/itzmeanjan/ette/app/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connect - Connecting to postgresql database
func Connect() *gorm.DB {
	_db, err := gorm.Open(postgres.Open(fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", cfg.Get("DB_USER"), cfg.Get("DB_PASSWORD"), cfg.Get("DB_HOST"), cfg.Get("DB_PORT"), cfg.Get("DB_NAME"))), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		log.Fatalf("[!] Failed to connect to db : %s\n", err.Error())
	}

	_db.AutoMigrate(&Blocks{}, &Transactions{}, &Events{})
	return _db
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

// GetTransaction - Fetches tx entry from database, given txhash & containing block hash
func GetTransaction(_db *gorm.DB, blkHash common.Hash, txHash common.Hash) *Transactions {
	var tx Transactions

	if err := _db.Where("hash = ? and blockhash = ?", txHash.Hex(), blkHash.Hex()).First(&tx).Error; err != nil {
		return nil
	}

	return &tx
}

// PutTransaction - Persisting transactions present in a block in database
func PutTransaction(_db *gorm.DB, _tx *types.Transaction, _txReceipt *types.Receipt, _sender common.Address) {
	var _pTx *Transactions

	// If tx creates contract, then we hold created contract address
	if _tx.To() == nil {
		_pTx = &Transactions{
			Hash:      _tx.Hash().Hex(),
			From:      _sender.Hex(),
			Contract:  _txReceipt.ContractAddress.Hex(),
			Gas:       _tx.Gas(),
			GasPrice:  _tx.GasPrice().String(),
			Cost:      _tx.Cost().String(),
			Nonce:     _tx.Nonce(),
			State:     _txReceipt.Status,
			BlockHash: _txReceipt.BlockHash.Hex(),
		}
	} else {
		// This is a normal tx, so we keep contract field empty
		_pTx = &Transactions{
			Hash:      _tx.Hash().Hex(),
			From:      _sender.Hex(),
			To:        _tx.To().Hex(),
			Gas:       _tx.Gas(),
			GasPrice:  _tx.GasPrice().String(),
			Cost:      _tx.Cost().String(),
			Nonce:     _tx.Nonce(),
			State:     _txReceipt.Status,
			BlockHash: _txReceipt.BlockHash.Hex(),
		}
	}

	if err := _db.Create(_pTx).Error; err != nil {
		log.Printf("[!] Failed to persist tx [ block : %s ] : %s\n", _txReceipt.BlockNumber.String(), err.Error())
	}
}

// CheckPersistanceStatusOfEvents - Given tx receipt, it finds out whether all log entries are persisted or not
func CheckPersistanceStatusOfEvents(_db *gorm.DB, _txReceipt *types.Receipt) bool {
	count := 0

	for _, v := range _txReceipt.Logs {
		var _event Events

		if err := _db.Where("index = ? and blockhash = ?", v.Index, v.BlockHash.Hex()).First(&_event).Error; err == nil && _event.Index == v.Index && _event.BlockHash == v.BlockHash.Hex() {
			count++
		}

	}

	return count == len(_txReceipt.Logs)
}

// PutEvent - Entering new log events emitted as result of execution of EVM transaction
// into persistable storage
func PutEvent(_db *gorm.DB, _txReceipt *types.Receipt) {
	for _, v := range _txReceipt.Logs {
		if err := _db.Create(&Events{
			Origin:          v.Address.Hex(),
			Index:           v.Index,
			Topics:          com.StringifyEventTopics(v.Topics),
			Data:            v.Data,
			TransactionHash: v.TxHash.Hex(),
			BlockHash:       v.BlockHash.Hex(),
		}).Error; err != nil {
			log.Printf("[!] Failed to persist tx log [ block : %s ] : %s\n", _txReceipt.BlockNumber.String(), err.Error())
		}
	}
}
