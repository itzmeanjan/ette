package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"encoding/binary"
	"errors"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	cmn "github.com/itzmeanjan/ette/app/common"
	cfg "github.com/itzmeanjan/ette/app/config"
	_db "github.com/itzmeanjan/ette/app/db"
	"github.com/itzmeanjan/ette/app/rest/graph/generated"
	"github.com/itzmeanjan/ette/app/rest/graph/model"
)

func (r *queryResolver) BlockByHash(ctx context.Context, hash string) (*model.Block, error) {
	if !(strings.HasPrefix(hash, "0x") && len(hash) == 66) {
		return nil, errors.New("Bad Block Hash")
	}

	return getGraphQLCompatibleBlock(ctx, _db.GetBlockByHash(db, common.HexToHash(hash)), true)
}

func (r *queryResolver) BlockByNumber(ctx context.Context, number string) (*model.Block, error) {
	_number, err := strconv.ParseUint(number, 10, 64)
	if err != nil {
		return nil, errors.New("Bad Block Number")
	}

	return getGraphQLCompatibleBlock(ctx, _db.GetBlockByNumber(db, _number), true)
}

func (r *queryResolver) BlocksByNumberRange(ctx context.Context, from string, to string) ([]*model.Block, error) {
	_from, _to, err := cmn.RangeChecker(from, to, cfg.GetBlockNumberRange())
	if err != nil {
		return nil, errors.New("Bad Block Number Range")
	}

	return getGraphQLCompatibleBlocks(ctx, _db.GetBlocksByNumberRange(db, _from, _to))
}

func (r *queryResolver) BlocksByTimeRange(ctx context.Context, from string, to string) ([]*model.Block, error) {
	_from, _to, err := cmn.RangeChecker(from, to, cfg.GetTimeRange())
	if err != nil {
		return nil, errors.New("Bad Block Timestamp Range")
	}

	return getGraphQLCompatibleBlocks(ctx, _db.GetBlocksByTimeRange(db, _from, _to))
}

func (r *queryResolver) Transaction(ctx context.Context, hash string) (*model.Transaction, error) {
	if !(strings.HasPrefix(hash, "0x") && len(hash) == 66) {
		return nil, errors.New("Bad Transaction Hash")
	}

	return getGraphQLCompatibleTransaction(ctx, _db.GetTransactionByHash(db, common.HexToHash(hash)), true)
}

func (r *queryResolver) TransactionCountByBlockHash(ctx context.Context, hash string) (int, error) {
	if !(strings.HasPrefix(hash, "0x") && len(hash) == 66) {
		return 0, errors.New("Bad Block Hash")
	}

	count := int(_db.GetTransactionCountByBlockHash(db, common.HexToHash(hash)))

	// Attempting to calculate byte form of number
	// so that we can keep track of how much data was transferred
	// to client
	_count := make([]byte, 4)
	binary.LittleEndian.PutUint32(_count, uint32(count))

	if err := doBookKeeping(ctx, _count); err != nil {
		return 0, errors.New("Book keeping failed")
	}

	return count, nil
}

func (r *queryResolver) TransactionsByBlockHash(ctx context.Context, hash string) ([]*model.Transaction, error) {
	if !(strings.HasPrefix(hash, "0x") && len(hash) == 66) {
		return nil, errors.New("Bad Block Hash")
	}

	return getGraphQLCompatibleTransactions(ctx, _db.GetTransactionsByBlockHash(db, common.HexToHash(hash)))
}

func (r *queryResolver) TransactionCountByBlockNumber(ctx context.Context, number string) (int, error) {
	_number, err := strconv.ParseUint(number, 10, 64)
	if err != nil {
		return 0, errors.New("Bad Block Number")
	}

	count := int(_db.GetTransactionCountByBlockNumber(db, _number))

	// Attempting to calculate byte form of number
	// so that we can keep track of how much data was transferred
	// to client
	_count := make([]byte, 4)
	binary.LittleEndian.PutUint32(_count, uint32(count))

	if err := doBookKeeping(ctx, _count); err != nil {
		return 0, errors.New("Book keeping failed")
	}

	return count, nil
}

func (r *queryResolver) TransactionsByBlockNumber(ctx context.Context, number string) ([]*model.Transaction, error) {
	_number, err := strconv.ParseUint(number, 10, 64)
	if err != nil {
		return nil, errors.New("Bad Block Number")
	}

	return getGraphQLCompatibleTransactions(ctx, _db.GetTransactionsByBlockNumber(db, _number))
}

func (r *queryResolver) TransactionCountFromAccountByNumberRange(ctx context.Context, account string, from string, to string) (int, error) {
	if !(strings.HasPrefix(account, "0x") && len(account) == 42) {
		return 0, errors.New("Bad Account Address")
	}

	_from, _to, err := cmn.RangeChecker(from, to, cfg.GetBlockNumberRange())
	if err != nil {
		return 0, errors.New("Bad Block Number Range")
	}

	count := int(_db.GetTransactionCountFromAccountByBlockNumberRange(db, common.HexToAddress(account), _from, _to))

	// Attempting to calculate byte form of number
	// so that we can keep track of how much data was transferred
	// to client
	_count := make([]byte, 4)
	binary.LittleEndian.PutUint32(_count, uint32(count))

	if err := doBookKeeping(ctx, _count); err != nil {
		return 0, errors.New("Book keeping failed")
	}

	return count, nil
}

func (r *queryResolver) TransactionsFromAccountByNumberRange(ctx context.Context, account string, from string, to string) ([]*model.Transaction, error) {
	if !(strings.HasPrefix(account, "0x") && len(account) == 42) {
		return nil, errors.New("Bad Account Address")
	}

	_from, _to, err := cmn.RangeChecker(from, to, cfg.GetBlockNumberRange())
	if err != nil {
		return nil, errors.New("Bad Block Number Range")
	}

	return getGraphQLCompatibleTransactions(ctx, _db.GetTransactionsFromAccountByBlockNumberRange(db, common.HexToAddress(account), _from, _to))
}

func (r *queryResolver) TransactionCountFromAccountByTimeRange(ctx context.Context, account string, from string, to string) (int, error) {
	if !(strings.HasPrefix(account, "0x") && len(account) == 42) {
		return 0, errors.New("Bad Account Address")
	}

	_from, _to, err := cmn.RangeChecker(from, to, cfg.GetTimeRange())
	if err != nil {
		return 0, errors.New("Bad Block Timestamp Range")
	}

	count := int(_db.GetTransactionCountFromAccountByBlockTimeRange(db, common.HexToAddress(account), _from, _to))

	// Attempting to calculate byte form of number
	// so that we can keep track of how much data was transferred
	// to client
	_count := make([]byte, 4)
	binary.LittleEndian.PutUint32(_count, uint32(count))

	if err := doBookKeeping(ctx, _count); err != nil {
		return 0, errors.New("Book keeping failed")
	}

	return count, nil
}

func (r *queryResolver) TransactionsFromAccountByTimeRange(ctx context.Context, account string, from string, to string) ([]*model.Transaction, error) {
	if !(strings.HasPrefix(account, "0x") && len(account) == 42) {
		return nil, errors.New("Bad Account Address")
	}

	_from, _to, err := cmn.RangeChecker(from, to, cfg.GetTimeRange())
	if err != nil {
		return nil, errors.New("Bad Block Timestamp Range")
	}

	return getGraphQLCompatibleTransactions(ctx, _db.GetTransactionsFromAccountByBlockTimeRange(db, common.HexToAddress(account), _from, _to))
}

func (r *queryResolver) TransactionCountToAccountByNumberRange(ctx context.Context, account string, from string, to string) (int, error) {
	if !(strings.HasPrefix(account, "0x") && len(account) == 42) {
		return 0, errors.New("Bad Account Address")
	}

	_from, _to, err := cmn.RangeChecker(from, to, cfg.GetBlockNumberRange())
	if err != nil {
		return 0, errors.New("Bad Block Number Range")
	}

	count := int(_db.GetTransactionCountToAccountByBlockNumberRange(db, common.HexToAddress(account), _from, _to))

	// Attempting to calculate byte form of number
	// so that we can keep track of how much data was transferred
	// to client
	_count := make([]byte, 4)
	binary.LittleEndian.PutUint32(_count, uint32(count))

	if err := doBookKeeping(ctx, _count); err != nil {
		return 0, errors.New("Book keeping failed")
	}

	return count, nil
}

func (r *queryResolver) TransactionsToAccountByNumberRange(ctx context.Context, account string, from string, to string) ([]*model.Transaction, error) {
	if !(strings.HasPrefix(account, "0x") && len(account) == 42) {
		return nil, errors.New("Bad Account Address")
	}

	_from, _to, err := cmn.RangeChecker(from, to, cfg.GetBlockNumberRange())
	if err != nil {
		return nil, errors.New("Bad Block Number Range")
	}

	return getGraphQLCompatibleTransactions(ctx, _db.GetTransactionsToAccountByBlockNumberRange(db, common.HexToAddress(account), _from, _to))
}

func (r *queryResolver) TransactionCountToAccountByTimeRange(ctx context.Context, account string, from string, to string) (int, error) {
	if !(strings.HasPrefix(account, "0x") && len(account) == 42) {
		return 0, errors.New("Bad Account Address")
	}

	_from, _to, err := cmn.RangeChecker(from, to, cfg.GetTimeRange())
	if err != nil {
		return 0, errors.New("Bad Block Timestamp Range")
	}

	count := int(_db.GetTransactionCountToAccountByBlockTimeRange(db, common.HexToAddress(account), _from, _to))

	// Attempting to calculate byte form of number
	// so that we can keep track of how much data was transferred
	// to client
	_count := make([]byte, 4)
	binary.LittleEndian.PutUint32(_count, uint32(count))

	if err := doBookKeeping(ctx, _count); err != nil {
		return 0, errors.New("Book keeping failed")
	}

	return count, nil
}

func (r *queryResolver) TransactionsToAccountByTimeRange(ctx context.Context, account string, from string, to string) ([]*model.Transaction, error) {
	if !(strings.HasPrefix(account, "0x") && len(account) == 42) {
		return nil, errors.New("Bad Account Address")
	}

	_from, _to, err := cmn.RangeChecker(from, to, cfg.GetTimeRange())
	if err != nil {
		return nil, errors.New("Bad Block Timestamp Range")
	}

	return getGraphQLCompatibleTransactions(ctx, _db.GetTransactionsToAccountByBlockTimeRange(db, common.HexToAddress(account), _from, _to))
}

func (r *queryResolver) TransactionCountBetweenAccountsByNumberRange(ctx context.Context, fromAccount string, toAccount string, from string, to string) (int, error) {
	if !(strings.HasPrefix(fromAccount, "0x") && len(fromAccount) == 42) {
		return 0, errors.New("Bad From Account Address")
	}

	if !(strings.HasPrefix(toAccount, "0x") && len(toAccount) == 42) {
		return 0, errors.New("Bad To Account Address")
	}

	_from, _to, err := cmn.RangeChecker(from, to, cfg.GetBlockNumberRange())
	if err != nil {
		return 0, errors.New("Bad Block Number Range")
	}

	count := int(_db.GetTransactionCountBetweenAccountsByBlockNumberRange(db, common.HexToAddress(fromAccount), common.HexToAddress(toAccount), _from, _to))

	// Attempting to calculate byte form of number
	// so that we can keep track of how much data was transferred
	// to client
	_count := make([]byte, 4)
	binary.LittleEndian.PutUint32(_count, uint32(count))

	if err := doBookKeeping(ctx, _count); err != nil {
		return 0, errors.New("Book keeping failed")
	}

	return count, nil
}

func (r *queryResolver) TransactionsBetweenAccountsByNumberRange(ctx context.Context, fromAccount string, toAccount string, from string, to string) ([]*model.Transaction, error) {
	if !(strings.HasPrefix(fromAccount, "0x") && len(fromAccount) == 42) {
		return nil, errors.New("Bad From Account Address")
	}

	if !(strings.HasPrefix(toAccount, "0x") && len(toAccount) == 42) {
		return nil, errors.New("Bad To Account Address")
	}

	_from, _to, err := cmn.RangeChecker(from, to, cfg.GetBlockNumberRange())
	if err != nil {
		return nil, errors.New("Bad Block Number Range")
	}

	return getGraphQLCompatibleTransactions(ctx, _db.GetTransactionsBetweenAccountsByBlockNumberRange(db, common.HexToAddress(fromAccount), common.HexToAddress(toAccount), _from, _to))
}

func (r *queryResolver) TransactionCountBetweenAccountsByTimeRange(ctx context.Context, fromAccount string, toAccount string, from string, to string) (int, error) {
	if !(strings.HasPrefix(fromAccount, "0x") && len(fromAccount) == 42) {
		return 0, errors.New("Bad From Account Address")
	}

	if !(strings.HasPrefix(toAccount, "0x") && len(toAccount) == 42) {
		return 0, errors.New("Bad To Account Address")
	}

	_from, _to, err := cmn.RangeChecker(from, to, cfg.GetTimeRange())
	if err != nil {
		return 0, errors.New("Bad Block Timestamp Range")
	}

	count := int(_db.GetTransactionCountBetweenAccountsByBlockTimeRange(db, common.HexToAddress(fromAccount), common.HexToAddress(toAccount), _from, _to))

	// Attempting to calculate byte form of number
	// so that we can keep track of how much data was transferred
	// to client
	_count := make([]byte, 4)
	binary.LittleEndian.PutUint32(_count, uint32(count))

	if err := doBookKeeping(ctx, _count); err != nil {
		return 0, errors.New("Book keeping failed")
	}

	return count, nil
}

func (r *queryResolver) TransactionsBetweenAccountsByTimeRange(ctx context.Context, fromAccount string, toAccount string, from string, to string) ([]*model.Transaction, error) {
	if !(strings.HasPrefix(fromAccount, "0x") && len(fromAccount) == 42) {
		return nil, errors.New("Bad From Account Address")
	}

	if !(strings.HasPrefix(toAccount, "0x") && len(toAccount) == 42) {
		return nil, errors.New("Bad To Account Address")
	}

	_from, _to, err := cmn.RangeChecker(from, to, cfg.GetTimeRange())
	if err != nil {
		return nil, errors.New("Bad Block Timestamp Range")
	}

	return getGraphQLCompatibleTransactions(ctx, _db.GetTransactionsBetweenAccountsByBlockTimeRange(db, common.HexToAddress(fromAccount), common.HexToAddress(toAccount), _from, _to))
}

func (r *queryResolver) ContractsCreatedFromAccountByNumberRange(ctx context.Context, account string, from string, to string) ([]*model.Transaction, error) {
	if !(strings.HasPrefix(account, "0x") && len(account) == 42) {
		return nil, errors.New("Bad Account Address")
	}

	_from, _to, err := cmn.RangeChecker(from, to, cfg.GetBlockNumberRange())
	if err != nil {
		return nil, errors.New("Bad Block Number Range")
	}

	return getGraphQLCompatibleTransactions(ctx, _db.GetContractCreationTransactionsFromAccountByBlockNumberRange(db, common.HexToAddress(account), _from, _to))
}

func (r *queryResolver) ContractsCreatedFromAccountByTimeRange(ctx context.Context, account string, from string, to string) ([]*model.Transaction, error) {
	if !(strings.HasPrefix(account, "0x") && len(account) == 42) {
		return nil, errors.New("Bad Account Address")
	}

	_from, _to, err := cmn.RangeChecker(from, to, cfg.GetTimeRange())
	if err != nil {
		return nil, errors.New("Bad Block Timestamp Range")
	}

	return getGraphQLCompatibleTransactions(ctx, _db.GetContractCreationTransactionsFromAccountByBlockTimeRange(db, common.HexToAddress(account), _from, _to))
}

func (r *queryResolver) TransactionFromAccountWithNonce(ctx context.Context, account string, nonce string) (*model.Transaction, error) {
	if !(strings.HasPrefix(account, "0x") && len(account) == 42) {
		return nil, errors.New("Bad Account Address")
	}

	_nonce, err := strconv.ParseUint(nonce, 10, 64)
	if err != nil {
		return nil, errors.New("Bad Account Nonce")
	}

	return getGraphQLCompatibleTransaction(ctx, _db.GetTransactionFromAccountWithNonce(db, common.HexToAddress(account), _nonce), true)
}

func (r *queryResolver) EventsFromContractByNumberRange(ctx context.Context, contract string, from string, to string) ([]*model.Event, error) {
	if !(strings.HasPrefix(contract, "0x") && len(contract) == 42) {
		return nil, errors.New("Bad Contract Address")
	}

	_from, _to, err := cmn.RangeChecker(from, to, cfg.GetBlockNumberRange())
	if err != nil {
		return nil, errors.New("Bad Block Number Range")
	}

	return getGraphQLCompatibleEvents(ctx, _db.GetEventsFromContractByBlockNumberRange(db, common.HexToAddress(contract), _from, _to))
}

func (r *queryResolver) EventsFromContractByTimeRange(ctx context.Context, contract string, from string, to string) ([]*model.Event, error) {
	if !(strings.HasPrefix(contract, "0x") && len(contract) == 42) {
		return nil, errors.New("Bad Contract Address")
	}

	_from, _to, err := cmn.RangeChecker(from, to, cfg.GetTimeRange())
	if err != nil {
		return nil, errors.New("Bad Block Timestamp Range")
	}

	return getGraphQLCompatibleEvents(ctx, _db.GetEventsFromContractByBlockTimeRange(db, common.HexToAddress(contract), _from, _to))
}

func (r *queryResolver) EventsByBlockHash(ctx context.Context, hash string) ([]*model.Event, error) {
	if !(strings.HasPrefix(hash, "0x") && len(hash) == 66) {
		return nil, errors.New("Bad Block Hash")
	}

	return getGraphQLCompatibleEvents(ctx, _db.GetEventsByBlockHash(db, common.HexToHash(hash)))
}

func (r *queryResolver) EventsByTxHash(ctx context.Context, hash string) ([]*model.Event, error) {
	if !(strings.HasPrefix(hash, "0x") && len(hash) == 66) {
		return nil, errors.New("Bad Transaction Hash")
	}

	return getGraphQLCompatibleEvents(ctx, _db.GetEventsByTransactionHash(db, common.HexToHash(hash)))
}

func (r *queryResolver) EventsFromContractWithTopicsByNumberRange(ctx context.Context, contract string, from string, to string, topics []string) ([]*model.Event, error) {
	if !(strings.HasPrefix(contract, "0x") && len(contract) == 42) {
		return nil, errors.New("Bad Contract Address")
	}

	_from, _to, err := cmn.RangeChecker(from, to, cfg.GetBlockNumberRange())
	if err != nil {
		return nil, errors.New("Bad Block Number Range")
	}

	return getGraphQLCompatibleEvents(ctx, _db.GetEventsFromContractWithTopicsByBlockNumberRange(db, common.HexToAddress(contract), _from, _to, cmn.CreateEventTopicMap(FillUpTopicArray(topics))))
}

func (r *queryResolver) EventsFromContractWithTopicsByTimeRange(ctx context.Context, contract string, from string, to string, topics []string) ([]*model.Event, error) {
	if !(strings.HasPrefix(contract, "0x") && len(contract) == 42) {
		return nil, errors.New("Bad Contract Address")
	}

	_from, _to, err := cmn.RangeChecker(from, to, cfg.GetTimeRange())
	if err != nil {
		return nil, errors.New("Bad Block Timestamp Range")
	}

	return getGraphQLCompatibleEvents(ctx, _db.GetEventsFromContractWithTopicsByBlockTimeRange(db, common.HexToAddress(contract), _from, _to, cmn.CreateEventTopicMap(FillUpTopicArray(topics))))
}

func (r *queryResolver) LastXEventsFromContract(ctx context.Context, contract string, x int) ([]*model.Event, error) {
	if !(strings.HasPrefix(contract, "0x") && len(contract) == 42) {
		return nil, errors.New("Bad Contract Address")
	}

	if !(x <= 50) {
		return nil, errors.New("Too Many Events Requested")
	}

	return getGraphQLCompatibleEvents(ctx, _db.GetLastXEventsFromContract(db, common.HexToAddress(contract), x))
}

func (r *queryResolver) EventByBlockHashAndLogIndex(ctx context.Context, hash string, index string) (*model.Event, error) {
	if !(strings.HasPrefix(hash, "0x") && len(hash) == 66) {
		return nil, errors.New("Bad Block Hash")
	}

	_index, err := strconv.ParseUint(index, 10, 64)
	if err != nil {
		return nil, errors.New("Bad Log Index")
	}

	return getGraphQLCompatibleEvent(ctx, _db.GetEventByBlockHashAndLogIndex(db, common.HexToHash(hash), uint(_index)), true)
}

func (r *queryResolver) EventByBlockNumberAndLogIndex(ctx context.Context, number string, index string) (*model.Event, error) {
	_number, err := strconv.ParseUint(number, 10, 64)
	if err != nil {
		return nil, errors.New("Bad Block Number")
	}

	_index, err := strconv.ParseUint(index, 10, 64)
	if err != nil {
		return nil, errors.New("Bad Log Index")
	}

	return getGraphQLCompatibleEvent(ctx, _db.GetEventByBlockNumberAndLogIndex(db, _number, uint(_index)), true)
}

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }

// !!! WARNING !!!
// The code below was going to be deleted when updating resolvers. It has been copied here so you have
// one last chance to move it out of harms way if you want. There are two reasons this happens:
//  - When renaming or deleting a resolver the old code will be put in here. You can safely delete
//    it when you're done.
//  - You have helper methods in this file. Move them out to keep these resolver files clean.
func (r *queryResolver) TransactionByHash(ctx context.Context, hash string) (*model.Transaction, error) {
	if !(strings.HasPrefix(hash, "0x") && len(hash) == 66) {
		return nil, errors.New("Bad Transaction Hash")
	}

	return getGraphQLCompatibleTransaction(ctx, _db.GetTransactionByHash(db, common.HexToHash(hash)), true)
}
