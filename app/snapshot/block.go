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
		TransactionRootHash: block.TransactionRootHash,
		ReceiptRootHash:     block.ReceiptRootHash,
	}

	txs := _db.GetTransactionsByBlockHash(db, common.HexToHash(block.Hash))
	if txs == nil {
		return _block
	}

	_block.Transactions = TransactionsToProtoBuf(txs, db)
	return _block

}

// BlocksToProtoBuf - Creating proto buffer compatible data
// format for blocks data, which can be easily serialized & deserialized
// for taking snapshot and restoring from it
func BlocksToProtoBuf(blocks *data.Blocks, db *gorm.DB) *pb.Blocks {

	_blocks := &pb.Blocks{
		Blocks: make([]*pb.Block, len(blocks.Blocks)),
	}

	for i := 0; i < len(blocks.Blocks); i++ {
		_blocks.Blocks[i] = BlockToProtoBuf(blocks.Blocks[i], db)
	}

	return _blocks

}
