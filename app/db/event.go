package db

import (
	"errors"

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

// StoreEvent - Either creating new event entry or updating existing one
// depending upon whether entry exists already or not
//
// This may be also case when nothing of 👆 is performed, because already
// we've updated data
//
// When reading from database, query not to be wrapped inside any transaction
// but while making any change, it'll be definitely protected using transaction
// to make whole block persistance operation little bit atomic
func StoreEvent(dbWOTx *gorm.DB, dbWTx *gorm.DB, event *Events) error {

	if event == nil {
		return errors.New("Empty event received while attempting to persist")
	}

	persistedEvent := GetEvent(dbWOTx, event.Index, event.BlockHash)
	if persistedEvent == nil {
		return PutEvent(dbWTx, event)
	}

	if !persistedEvent.SimilarTo(event) {
		return UpdateEvent(dbWTx, event)
	}

	return nil

}

// GetEvent - Given event index in block & block hash, returns event which is
// matching from database
func GetEvent(_db *gorm.DB, index uint, blockHash string) *Events {
	var event Events

	if err := _db.Where("index = ? and blockhash = ?", index, blockHash).First(&event).Error; err != nil {
		return nil
	}

	return &event
}

// PutEvent - Persists event log into database
func PutEvent(tx *gorm.DB, event *Events) error {
	return tx.Create(event).Error
}

// UpdateEvent - Updating already existing event data
func UpdateEvent(tx *gorm.DB, event *Events) error {
	return tx.Where("index = ? and blockhash = ?", event.Index, event.BlockHash).Updates(event).Error
}
