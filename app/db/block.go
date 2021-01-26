package db

import (
	"errors"

	d "github.com/itzmeanjan/ette/app/data"
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
func StoreBlock(dbWOTx *gorm.DB, block *PackedBlock, status *d.StatusHolder) error {

	if block == nil {
		return errors.New("Empty block received while attempting to persist")
	}

	// -- Starting DB transaction
	return dbWOTx.Transaction(func(dbWTx *gorm.DB) error {

		blockInserted := false

		persistedBlock := GetBlock(dbWOTx, block.Block.Number)
		if persistedBlock == nil {

			if err := PutBlock(dbWTx, block.Block); err != nil {
				return err
			}

			blockInserted = true

		} else if !persistedBlock.SimilarTo(block.Block) {

			if err := UpdateBlock(dbWTx, block.Block); err != nil {
				return err
			}

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
		if blockInserted && status != nil {
			status.IncrementBlocksInserted()
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
func PutBlock(tx *gorm.DB, block *Blocks) error {
	return tx.Create(block).Error
}

// UpdateBlock - Updating already existing block
func UpdateBlock(tx *gorm.DB, block *Blocks) error {
	return tx.Where("number = ?", block.Number).Updates(block).Error
}
