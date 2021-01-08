package db

import (
	"errors"

	"gorm.io/gorm"
)

// StoreTransaction - Stores specific tx into database, where presence of tx
// in database being performed using db handle ( i.e. `dbWOTx` ) which is not
// wrapped inside database transaction, but whenever performing any writing to database
// uses db handle which is always passed to db engine, wrapped inside a db transaction
// i.e. `dbWTx`, which protects helps us in rolling back to previous state, in case of some failure
// faced during this persistance stage
func StoreTransaction(dbWOTx *gorm.DB, dbWTx *gorm.DB, tx *Transactions) error {

	if tx == nil {
		return errors.New("Empty transaction received while attempting to persist")
	}

	persistedTx := GetTransaction(dbWOTx, tx.BlockHash, tx.BlockHash)
	if persistedTx == nil {
		return PutTransaction(dbWTx, tx)
	}

	if !persistedTx.SimilarTo(tx) {
		return UpdateTransaction(dbWTx, tx)
	}

	return nil

}

// GetTransaction - Fetches tx entry from database, given txhash & containing block hash
func GetTransaction(_db *gorm.DB, blkHash string, txHash string) *Transactions {
	var tx Transactions

	if err := _db.Where("hash = ? and blockhash = ?", txHash, blkHash).First(&tx).Error; err != nil {
		return nil
	}

	return &tx
}

// PutTransaction - Persisting transaction
func PutTransaction(tx *gorm.DB, txn *Transactions) error {
	return tx.Create(txn).Error
}

// UpdateTransaction - Updating already persisted transaction
func UpdateTransaction(tx *gorm.DB, txn *Transactions) error {
	return tx.Where("hash = ? and blockhash = ?", txn.Hash, txn.BlockHash).Updates(txn).Error
}
