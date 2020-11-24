package db

import (
	"log"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"gorm.io/gorm"
)

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
