package db

import (
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// UpsertEvent - It may be the case previously this block was processed
// and event log also got persisted into database, which has been updated
// on chain, due to chain reorganization
//
// Now we want to persist latest entry into database, using upsert operation
// i.e. if entry doesn't exist yet, it'll be created, but if it does
// i.e. conflicting primary key found, then, all fields will be updated to latest value
func UpsertEvent(dbWTx *gorm.DB, event *Events) error {

	if event == nil {
		return errors.New("Empty event received while attempting to persist")
	}

	return dbWTx.Clauses(clause.OnConflict{UpdateAll: true}).Create(event).Error

}

// RemoveEventsByBlockHash - All events emitted by tx(s) packed in block, to be
// removed from DB
func RemoveEventsByBlockHash(dbWTx *gorm.DB, blockHash common.Hash) error {

	return dbWTx.Delete(&Events{BlockHash: blockHash.Hex()}).Error

}
