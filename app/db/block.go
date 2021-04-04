package db

import (
	"errors"
	"log"

	d "github.com/itzmeanjan/ette/app/data"
	q "github.com/itzmeanjan/ette/app/queue"
	"gorm.io/gorm"
)

// StoreBlock - Persisting block data in database,
// if data already not stored
//
// Also checks equality with existing data, if mismatch found,
// updated with latest data
//
// Tries to wrap db modifications inside database transaction to
// guarantee consistency, other read only operations being performed without
// protection of db transaction
//
// 👆 gives us performance improvement, also taste of atomic db operation
// i.e. either whole block data is written or nothing is written
func StoreBlock(dbWOTx *gorm.DB, block *PackedBlock, status *d.StatusHolder, queue *q.BlockProcessorQueue) error {

	if block == nil {
		return errors.New("empty block received while attempting to persist")
	}

	// -- Starting DB transaction
	return dbWOTx.Transaction(func(dbWTx *gorm.DB) error {

		blockInserted := false

		persistedBlock := GetBlock(dbWTx, block.Block.Number)
		if persistedBlock == nil {

			if err := PutBlock(dbWTx, block.Block); err != nil {
				return err
			}

			blockInserted = true

		} else if !persistedBlock.SimilarTo(block.Block) {

			log.Printf("[!] Block %d already present in DB, similar ❌\n", block.Block.Number)

			// -- If block is going to be updated, it's better
			// we also remove associated entries for that block
			// i.e. transactions, events
			if err := RemoveEventsByBlockHash(dbWTx, persistedBlock.Hash); err != nil {
				return err
			}

			if err := RemoveTransactionsByBlockHash(dbWTx, persistedBlock.Hash); err != nil {
				return err
			}
			// -- block data clean up ends here

			if err := UpdateBlock(dbWTx, block.Block); err != nil {
				return err
			}

		} else {

			log.Printf("[+] Block %d already present in DB, similar ✅\n", block.Block.Number)
			return nil

		}

		if block.Transactions == nil {

			// During 👆 flow, if we've really inserted a new block into database,
			// count will get updated
			if blockInserted && status != nil {
				status.IncrementBlocksInserted()
			}

			return nil

		}

		for _, t := range block.Transactions {

			if err := UpsertTransaction(dbWTx, t.Tx); err != nil {
				return err
			}

			for _, e := range t.Events {

				if err := UpsertEvent(dbWTx, e); err != nil {
					return err
				}

			}

		}

		// During 👆 flow, if we've really inserted a new block into database,
		// count will get updated
		if blockInserted && status != nil && queue != nil {
			status.IncrementBlocksInserted() // @note This is to be removed
			queue.Inserted(block.Block.Number)
		}

		return nil

	})
	// -- Ending DB transaction

}

// GetBlock - Fetch block by number, from database
func GetBlock(_db *gorm.DB, number uint64) *Blocks {
	var block Blocks

	if err := _db.Where("number = ?", number).First(&block).Error; err != nil {
		return nil
	}

	return &block
}

// PutBlock - Persisting fetched block
func PutBlock(dbWTx *gorm.DB, block *Blocks) error {

	return dbWTx.Create(block).Error

}

// UpdateBlock - Updating already existing block
func UpdateBlock(dbWTx *gorm.DB, block *Blocks) error {

	return dbWTx.Model(&Blocks{}).Where("number = ?", block.Number).Updates(map[string]interface{}{
		"hash":            block.Hash,
		"time":            block.Time,
		"parenthash":      block.ParentHash,
		"difficulty":      block.Difficulty,
		"gasused":         block.GasUsed,
		"gaslimit":        block.GasLimit,
		"nonce":           block.Nonce,
		"miner":           block.Miner,
		"size":            block.Size,
		"stateroothash":   block.StateRootHash,
		"unclehash":       block.UncleHash,
		"txroothash":      block.TransactionRootHash,
		"receiptroothash": block.ReceiptRootHash,
		"extradata":       block.ExtraData,
	}).Error

}
