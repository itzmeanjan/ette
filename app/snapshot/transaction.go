package snapshot

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/itzmeanjan/ette/app/data"
	_db "github.com/itzmeanjan/ette/app/db"
	pb "github.com/itzmeanjan/ette/app/pb"
	"gorm.io/gorm"
)

// TransactionToProtoBuf - Creating proto buffer compatible data
// format for transaction data, which can be easily serialized & deserialized
// for taking snapshot and restoring from it
func TransactionToProtoBuf(tx *data.Transaction, db *gorm.DB) *pb.Transaction {

	_tx := &pb.Transaction{
		Hash:      tx.Hash,
		From:      tx.From,
		To:        tx.To,
		Contract:  tx.Contract,
		Value:     tx.Contract,
		Data:      tx.Data,
		Gas:       tx.Gas,
		GasPrice:  tx.GasPrice,
		Cost:      tx.Cost,
		Nonce:     tx.Nonce,
		State:     tx.State,
		BlockHash: tx.BlockHash,
	}

	events := _db.GetEventsByTransactionHash(db, common.HexToHash(tx.Hash))
	if events == nil {
		return _tx
	}

	_tx.Events = EventsToProtoBuf(events)
	return _tx

}

// TransactionsToProtoBuf - Creating proto buffer compatible data
// format for transactions data, which can be easily serialized & deserialized
// for taking snapshot and restoring from it
func TransactionsToProtoBuf(txs *data.Transactions, db *gorm.DB) []*pb.Transaction {

	_txs := make([]*pb.Transaction, len(txs.Transactions))

	for i := 0; i < len(txs.Transactions); i++ {
		_txs[i] = TransactionToProtoBuf(txs.Transactions[i], db)
	}

	return _txs

}

// ProtoBufToTransaction - Required while restoring from snapshot i.e. attempting to put
// whole block data into database
func ProtoBufToTransaction(tx *pb.Transaction) *_db.PackedTransaction {

	_tx := &_db.Transactions{
		Hash:      tx.Hash,
		From:      tx.From,
		To:        tx.To,
		Contract:  tx.Contract,
		Value:     tx.Contract,
		Data:      tx.Data,
		Gas:       tx.Gas,
		GasPrice:  tx.GasPrice,
		Cost:      tx.Cost,
		Nonce:     tx.Nonce,
		State:     tx.State,
		BlockHash: tx.BlockHash,
	}

	if tx.Events == nil {
		return &_db.PackedTransaction{
			Tx: _tx,
		}
	}

	return &_db.PackedTransaction{
		Tx:     _tx,
		Events: ProtoBufToEvents(tx.Events),
	}

}

// ProtoBufToTransactions - Required while restoring from snapshot i.e. attempting to put
// whole block data into database
func ProtoBufToTransactions(txs []*pb.Transaction) []*_db.PackedTransaction {

	_txs := make([]*_db.PackedTransaction, len(txs))

	for k, v := range txs {

		_txs[k] = ProtoBufToTransaction(v)

	}

	return _txs

}
