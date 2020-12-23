package db

import (
	"log"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	com "github.com/itzmeanjan/ette/app/common"
	"gorm.io/gorm"
)

// StoreEvents - Putting event data obtained from one tx
// into database, if not existing already
//
// Otherwise, it'll match with existing entry and decide
// whether updation is required or not
func StoreEvents(_db *gorm.DB, _txReceipt *types.Receipt) {
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

		persistedEvent := GetEvent(_db, v.Index, v.BlockHash)
		if persistedEvent == nil {
			PutEvent(_db, newEvent)
		}

		if !persistedEvent.SimilarTo(newEvent) {
			UpdateEvent(_db, newEvent)
		}

	}
}

// GetEvent - Given event index in block & block hash, returns event which is
// matching from database
func GetEvent(_db *gorm.DB, index uint, blockHash common.Hash) *Events {
	var event Events

	if err := _db.Where("index = ? and blockhash = ?", index, blockHash.Hex()).First(&event).Error; err != nil {
		return nil
	}

	return &event
}

// PutEvent - Persists event log into database
func PutEvent(_db *gorm.DB, event *Events) {

	if err := _db.Create(event).Error; err != nil {
		log.Printf("[!] Failed to persist tx log : %s\n", err.Error())
	}

}

// UpdateEvent - Updating event, already persisted in database,
// with latest info received
func UpdateEvent(_db *gorm.DB, event *Events) {

	if err := _db.Where("index = ? and blockhash = ?", event.Index, event.BlockHash).Updates(event).Error; err != nil {
		log.Printf("[!] Failed to update tx log : %s\n", err.Error())
	}

}
