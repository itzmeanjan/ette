package db

import (
	"log"

	"github.com/ethereum/go-ethereum/core/types"
	com "github.com/itzmeanjan/ette/app/common"
	"gorm.io/gorm"
)

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
