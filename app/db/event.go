package db

import (
	"github.com/ethereum/go-ethereum/core/types"
	com "github.com/itzmeanjan/ette/app/common"
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

// StoreEvents - Putting event data obtained from one tx
// into database, if not existing already
//
// Otherwise, it'll match with existing entry and decide
// whether updation is required or not
func StoreEvents(_db *gorm.DB, _txReceipt *types.Receipt) bool {
	count := 0

	for _, v := range _txReceipt.Logs {

		// Event which we're trying to persist
		// but it may already have been persisted
		//
		// So, we'll first check its existence in database
		// then try to match with this entry
		// and if not matching fully, we'll
		// simply try to update entry
		newEvent := &Events{
			Origin:          v.Address.Hex(),
			Index:           v.Index,
			Topics:          com.StringifyEventTopics(v.Topics),
			Data:            v.Data,
			TransactionHash: v.TxHash.Hex(),
			BlockHash:       v.BlockHash.Hex(),
		}

		persistedEvent := GetEvent(_db, v.Index, v.BlockHash.Hex())
		if persistedEvent == nil {
			if PutEvent(_db, newEvent) {
				count++
			}
			continue
		}

		if !persistedEvent.SimilarTo(newEvent) {
			if UpdateEvent(_db, newEvent) {
				count++
			}
			continue
		}

		count++
	}

	return count == len(_txReceipt.Logs)
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
