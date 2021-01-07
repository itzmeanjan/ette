package db

import (
	"gorm.io/gorm"
)

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
