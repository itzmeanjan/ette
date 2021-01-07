package db

import (
	"gorm.io/gorm"
)

// StoreTransaction - Stores specific tx into database, where presence of tx
// in database being performed using db handle ( i.e. `dbWOTx` ) which is not
// wrapped inside database transaction, but whenever performing any writing to database
// uses db handle which is always passed to db engine, wrapped inside a db transaction
// i.e. `dbWTx`, which protects helps us in rolling back to previous state, in case of some failure
// faced during this persistance stage
func StoreTransaction(dbWOTx *gorm.DB, dbWTx *gorm.DB, tx *Transactions) error {

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

// PutTransaction - Persisting transactions present in a block in database
func PutTransaction(tx *gorm.DB, txn *Transactions) error {
	if err := tx.Create(txn).Error; err != nil {
		return err
	}

	return nil
}

// UpdateTransaction - Updating already persisted transaction in database with
// new data
func UpdateTransaction(tx *gorm.DB, txn *Transactions) error {
	if err := tx.Where("hash = ? and blockhash = ?", txn.Hash, txn.BlockHash).Updates(txn).Error; err != nil {
		return err
	}

	return nil
}
