package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	_db "github.com/itzmeanjan/ette/app/db"
	"github.com/itzmeanjan/ette/app/rest/graph/generated"
	"github.com/itzmeanjan/ette/app/rest/graph/model"
)

func (r *queryResolver) BlockByHash(ctx context.Context, hash string) (*model.Block, error) {
	if !(strings.HasPrefix(hash, "0x") && len(hash) == 66) {
		return nil, errors.New("Bad Block Hash")
	}

	return getGraphQLCompatibleBlock(_db.GetBlockByHash(db, common.HexToHash(hash)))
}

func (r *queryResolver) BlockByNumber(ctx context.Context, number string) (*model.Block, error) {
	_number, err := strconv.ParseUint(number, 10, 64)
	if err != nil {
		return nil, errors.New("Bad Block Number")
	}

	return getGraphQLCompatibleBlock(_db.GetBlockByNumber(db, _number))
}

func (r *queryResolver) BlocksByNumberRange(ctx context.Context, from string, to string) ([]*model.Block, error) {
	_from, _to, err := rangeChecker(from, to, 10)
	if err != nil {
		return nil, errors.New("Bad Block Number Range")
	}

	_tmp := _db.GetBlocksByNumberRange(db, _from, _to)
	if _tmp == nil {
		return nil, errors.New("Found nothing")
	}

	return getGraphQLCompatibleBlocks(_tmp.Blocks)
}

func (r *queryResolver) BlocksByTimeRange(ctx context.Context, from string, to string) ([]*model.Block, error) {
	_from, _to, err := rangeChecker(from, to, 60)
	if err != nil {
		return nil, errors.New("Bad Block Timestamp Range")
	}

	_tmp := _db.GetBlocksByTimeRange(db, _from, _to)
	if _tmp == nil {
		return nil, errors.New("Found nothing")
	}

	return getGraphQLCompatibleBlocks(_tmp.Blocks)
}

func (r *queryResolver) TransactionsByBlockHash(ctx context.Context, hash string) ([]*model.Transaction, error) {
	if !(strings.HasPrefix(hash, "0x") && len(hash) == 66) {
		return nil, errors.New("Bad Block Hash")
	}

	_tmp := _db.GetTransactionsByBlockHash(db, common.HexToHash(hash))
	if _tmp == nil {
		return nil, errors.New("Found nothing")
	}

	return getGraphQLCompatibleTransactions(_tmp.Transactions)
}

func (r *queryResolver) TransactionsByBlockNumber(ctx context.Context, number string) ([]*model.Transaction, error) {
	_number, err := strconv.ParseUint(number, 10, 64)
	if err != nil {
		return nil, errors.New("Bad Block Number")
	}

	_tmp := _db.GetTransactionsByBlockNumber(db, _number)
	if _tmp == nil {
		return nil, errors.New("Found nothing")
	}

	return getGraphQLCompatibleTransactions(_tmp.Transactions)
}

func (r *queryResolver) Transaction(ctx context.Context, hash string) (*model.Transaction, error) {
	if !(strings.HasPrefix(hash, "0x") && len(hash) == 66) {
		return nil, errors.New("Bad Transaction Hash")
	}

	return getGraphQLCompatibleTransaction(_db.GetTransactionByHash(db, common.HexToHash(hash)))
}

func (r *queryResolver) TransactionsFromAccountByNumberRange(ctx context.Context, account string, from string, to string) ([]*model.Transaction, error) {
	if !(strings.HasPrefix(account, "0x") && len(account) == 42) {
		return nil, errors.New("Bad Account Address")
	}

	_from, _to, err := rangeChecker(from, to, 100)
	if err != nil {
		return nil, errors.New("Bad Block Number Range")
	}

	_tmp := _db.GetTransactionsFromAccountByBlockNumberRange(db, common.HexToAddress(account), _from, _to)
	if _tmp == nil {
		return nil, errors.New("Found nothing")
	}

	return getGraphQLCompatibleTransactions(_tmp.Transactions)
}

func (r *queryResolver) TransactionsFromAccountByTimeRange(ctx context.Context, account string, from string, to string) ([]*model.Transaction, error) {
	if !(strings.HasPrefix(account, "0x") && len(account) == 42) {
		return nil, errors.New("Bad Account Address")
	}

	_from, _to, err := rangeChecker(from, to, 600)
	if err != nil {
		return nil, errors.New("Bad Block Timestamp Range")
	}

	_tmp := _db.GetTransactionsFromAccountByBlockTimeRange(db, common.HexToAddress(account), _from, _to)
	if _tmp == nil {
		return nil, errors.New("Found nothing")
	}

	return getGraphQLCompatibleTransactions(_tmp.Transactions)
}

func (r *queryResolver) TransactionsToAccountByNumberRange(ctx context.Context, account string, from string, to string) ([]*model.Transaction, error) {
	if !(strings.HasPrefix(account, "0x") && len(account) == 42) {
		return nil, errors.New("Bad Account Address")
	}

	_from, _to, err := rangeChecker(from, to, 100)
	if err != nil {
		return nil, errors.New("Bad Block Number Range")
	}

	_tmp := _db.GetTransactionsToAccountByBlockNumberRange(db, common.HexToAddress(account), _from, _to)
	if _tmp == nil {
		return nil, errors.New("Found nothing")
	}

	return getGraphQLCompatibleTransactions(_tmp.Transactions)
}

func (r *queryResolver) TransactionsToAccountByTimeRange(ctx context.Context, account string, from string, to string) ([]*model.Transaction, error) {
	if !(strings.HasPrefix(account, "0x") && len(account) == 42) {
		return nil, errors.New("Bad Account Address")
	}

	_from, _to, err := rangeChecker(from, to, 600)
	if err != nil {
		return nil, errors.New("Bad Block Timestamp Range")
	}

	_tmp := _db.GetTransactionsToAccountByBlockTimeRange(db, common.HexToAddress(account), _from, _to)
	if _tmp == nil {
		return nil, errors.New("Found nothing")
	}

	return getGraphQLCompatibleTransactions(_tmp.Transactions)
}

func (r *queryResolver) TransactionsBetweenAccountsByNumberRange(ctx context.Context, fromAccount string, toAccount string, from string, to string) ([]*model.Transaction, error) {
	if !(strings.HasPrefix(fromAccount, "0x") && len(fromAccount) == 42) {
		return nil, errors.New("Bad From Account Address")
	}

	if !(strings.HasPrefix(toAccount, "0x") && len(toAccount) == 42) {
		return nil, errors.New("Bad To Account Address")
	}

	_from, _to, err := rangeChecker(from, to, 600)
	if err != nil {
		return nil, errors.New("Bad Block Number Range")
	}

	_tmp := _db.GetTransactionsBetweenAccountsByBlockNumberRange(db, common.HexToAddress(fromAccount), common.HexToAddress(toAccount), _from, _to)
	if _tmp == nil {
		return nil, errors.New("Found nothing")
	}

	return getGraphQLCompatibleTransactions(_tmp.Transactions)
}

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }
