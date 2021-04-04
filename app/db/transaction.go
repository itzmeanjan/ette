package db

import (
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// UpsertTransaction - It may be the case previously this block was processed
// and transaction also got persisted into database, which has been updated
// on chain, due to chain reorganization
//
// Now we want to persist latest entry into database, using upsert operation
// i.e. if entry doesn't exist yet, it'll be created, but if it does
// i.e. conflicting primary key found, then, all fields will be updated to latest value
func UpsertTransaction(dbWTx *gorm.DB, tx *Transactions) error {

	if tx == nil {
		return errors.New("empty transaction received while attempting to persist")
	}

	return dbWTx.Clauses(clause.OnConflict{UpdateAll: true}).Create(tx).Error

}

// RemoveTransactionsByBlockHash - Remove all tx(s) packed in this block from DB
func RemoveTransactionsByBlockHash(dbWTx *gorm.DB, blockHash string) error {

	return dbWTx.Where("blockhash = ?", blockHash).Delete(&Transactions{}).Error

}
