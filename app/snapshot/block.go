package snapshot

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/itzmeanjan/ette/app/data"
	_db "github.com/itzmeanjan/ette/app/db"
	pb "github.com/itzmeanjan/ette/app/pb"
	"gorm.io/gorm"
)

// BlockToProtoBuf - Creating proto buffer compatible data
// format for block data, which can be easily serialized & deserialized
// for taking snapshot and restoring from it
func BlockToProtoBuf(block *data.Block, db *gorm.DB) *pb.Block {

	_block := &pb.Block{
		Hash:                block.Hash,
		Number:              block.Number,
		Time:                block.Time,
		ParentHash:          block.ParentHash,
		Difficulty:          block.Difficulty,
		GasUsed:             block.GasUsed,
		GasLimit:            block.GasLimit,
		Nonce:               block.Nonce,
		Miner:               block.Miner,
		Size:                block.Size,
		StateRootHash:       block.StateRootHash,
		UncleHash:           block.UncleHash,
		TransactionRootHash: block.TransactionRootHash,
		ReceiptRootHash:     block.ReceiptRootHash,
		ExtraData:           block.ExtraData,
	}

	txs := _db.GetTransactionsByBlockHash(db, common.HexToHash(block.Hash))
	if txs == nil {
		return _block
	}

	_block.Transactions = TransactionsToProtoBuf(txs, db)
	return _block

}

// ProtoBufToBlock - Required while restoring from snapshot i.e. attempting to put
// whole block data into database
func ProtoBufToBlock(block *pb.Block) *_db.PackedBlock {

	_block := &_db.Blocks{
		Hash:                block.Hash,
		Number:              block.Number,
		Time:                block.Time,
		ParentHash:          block.ParentHash,
		Difficulty:          block.Difficulty,
		GasUsed:             block.GasUsed,
		GasLimit:            block.GasLimit,
		Nonce:               block.Nonce,
		Miner:               block.Miner,
		Size:                block.Size,
		StateRootHash:       block.StateRootHash,
		UncleHash:           block.UncleHash,
		TransactionRootHash: block.TransactionRootHash,
		ReceiptRootHash:     block.ReceiptRootHash,
		ExtraData:           block.ExtraData,
	}

	if block.Transactions == nil {
		return &_db.PackedBlock{
			Block: _block,
		}
	}

	return &_db.PackedBlock{
		Block:        _block,
		Transactions: ProtoBufToTransactions(block.Transactions),
	}

}
