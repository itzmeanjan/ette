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

	return getGraphQLCompatibleBlocks(_db.GetBlocksByNumberRange(db, _from, _to))
}

func (r *queryResolver) BlocksByTimeRange(ctx context.Context, from string, to string) ([]*model.Block, error) {
	_from, _to, err := rangeChecker(from, to, 60)
	if err != nil {
		return nil, errors.New("Bad Block Timestamp Range")
	}

	return getGraphQLCompatibleBlocks(_db.GetBlocksByTimeRange(db, _from, _to))
}

func (r *queryResolver) TransactionsByBlockHash(ctx context.Context, hash string) ([]*model.Transaction, error) {
	if !(strings.HasPrefix(hash, "0x") && len(hash) == 66) {
		return nil, errors.New("Bad Block Hash")
	}

	return getGraphQLCompatibleTransactions(_db.GetTransactionsByBlockHash(db, common.HexToHash(hash)))
}

func (r *queryResolver) TransactionsByBlockNumber(ctx context.Context, number string) ([]*model.Transaction, error) {
	_number, err := strconv.ParseUint(number, 10, 64)
	if err != nil {
		return nil, errors.New("Bad Block Number")
	}

	return getGraphQLCompatibleTransactions(_db.GetTransactionsByBlockNumber(db, _number))
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

	return getGraphQLCompatibleTransactions(_db.GetTransactionsFromAccountByBlockNumberRange(db, common.HexToAddress(account), _from, _to))
}

func (r *queryResolver) TransactionsFromAccountByTimeRange(ctx context.Context, account string, from string, to string) ([]*model.Transaction, error) {
	if !(strings.HasPrefix(account, "0x") && len(account) == 42) {
		return nil, errors.New("Bad Account Address")
	}

	_from, _to, err := rangeChecker(from, to, 600)
	if err != nil {
		return nil, errors.New("Bad Block Timestamp Range")
	}

	return getGraphQLCompatibleTransactions(_db.GetTransactionsFromAccountByBlockTimeRange(db, common.HexToAddress(account), _from, _to))
}

func (r *queryResolver) TransactionsToAccountByNumberRange(ctx context.Context, account string, from string, to string) ([]*model.Transaction, error) {
	if !(strings.HasPrefix(account, "0x") && len(account) == 42) {
		return nil, errors.New("Bad Account Address")
	}

	_from, _to, err := rangeChecker(from, to, 100)
	if err != nil {
		return nil, errors.New("Bad Block Number Range")
	}

	return getGraphQLCompatibleTransactions(_db.GetTransactionsToAccountByBlockNumberRange(db, common.HexToAddress(account), _from, _to))
}

func (r *queryResolver) TransactionsToAccountByTimeRange(ctx context.Context, account string, from string, to string) ([]*model.Transaction, error) {
	if !(strings.HasPrefix(account, "0x") && len(account) == 42) {
		return nil, errors.New("Bad Account Address")
	}

	_from, _to, err := rangeChecker(from, to, 600)
	if err != nil {
		return nil, errors.New("Bad Block Timestamp Range")
	}

	return getGraphQLCompatibleTransactions(_db.GetTransactionsToAccountByBlockTimeRange(db, common.HexToAddress(account), _from, _to))
}

func (r *queryResolver) TransactionsBetweenAccountsByNumberRange(ctx context.Context, fromAccount string, toAccount string, from string, to string) ([]*model.Transaction, error) {
	if !(strings.HasPrefix(fromAccount, "0x") && len(fromAccount) == 42) {
		return nil, errors.New("Bad From Account Address")
	}

	if !(strings.HasPrefix(toAccount, "0x") && len(toAccount) == 42) {
		return nil, errors.New("Bad To Account Address")
	}

	_from, _to, err := rangeChecker(from, to, 100)
	if err != nil {
		return nil, errors.New("Bad Block Number Range")
	}

	return getGraphQLCompatibleTransactions(_db.GetTransactionsBetweenAccountsByBlockNumberRange(db, common.HexToAddress(fromAccount), common.HexToAddress(toAccount), _from, _to))
}

func (r *queryResolver) TransactionsBetweenAccountsByTimeRange(ctx context.Context, fromAccount string, toAccount string, from string, to string) ([]*model.Transaction, error) {
	if !(strings.HasPrefix(fromAccount, "0x") && len(fromAccount) == 42) {
		return nil, errors.New("Bad From Account Address")
	}

	if !(strings.HasPrefix(toAccount, "0x") && len(toAccount) == 42) {
		return nil, errors.New("Bad To Account Address")
	}

	_from, _to, err := rangeChecker(from, to, 600)
	if err != nil {
		return nil, errors.New("Bad Block Timestamp Range")
	}

	return getGraphQLCompatibleTransactions(_db.GetTransactionsBetweenAccountsByBlockTimeRange(db, common.HexToAddress(fromAccount), common.HexToAddress(toAccount), _from, _to))
}

func (r *queryResolver) ContractsCreatedFromAccountByNumberRange(ctx context.Context, account string, from string, to string) ([]*model.Transaction, error) {
	if !(strings.HasPrefix(account, "0x") && len(account) == 42) {
		return nil, errors.New("Bad Account Address")
	}

	_from, _to, err := rangeChecker(from, to, 100)
	if err != nil {
		return nil, errors.New("Bad Block Number Range")
	}

	return getGraphQLCompatibleTransactions(_db.GetContractCreationTransactionsFromAccountByBlockNumberRange(db, common.HexToAddress(account), _from, _to))
}

func (r *queryResolver) ContractsCreatedFromAccountByTimeRange(ctx context.Context, account string, from string, to string) ([]*model.Transaction, error) {
	if !(strings.HasPrefix(account, "0x") && len(account) == 42) {
		return nil, errors.New("Bad Account Address")
	}

	_from, _to, err := rangeChecker(from, to, 600)
	if err != nil {
		return nil, errors.New("Bad Block Timestamp Range")
	}

	return getGraphQLCompatibleTransactions(_db.GetContractCreationTransactionsFromAccountByBlockTimeRange(db, common.HexToAddress(account), _from, _to))
}

func (r *queryResolver) TransactionFromAccountWithNonce(ctx context.Context, account string, nonce string) (*model.Transaction, error) {
	if !(strings.HasPrefix(account, "0x") && len(account) == 42) {
		return nil, errors.New("Bad Account Address")
	}

	_nonce, err := strconv.ParseUint(nonce, 10, 64)
	if err != nil {
		return nil, errors.New("Bad Account Nonce")
	}

	return getGraphQLCompatibleTransaction(_db.GetTransactionFromAccountWithNonce(db, common.HexToAddress(account), _nonce))
}

func (r *queryResolver) EventsFromContractByNumberRange(ctx context.Context, contract string, from string, to string) ([]*model.Event, error) {
	if !(strings.HasPrefix(contract, "0x") && len(contract) == 42) {
		return nil, errors.New("Bad Contract Address")
	}

	_from, _to, err := rangeChecker(from, to, 10)
	if err != nil {
		return nil, errors.New("Bad Block Number Range")
	}

	return getGraphQLCompatibleEvents(_db.GetEventsFromContractByBlockNumberRange(db, common.HexToAddress(contract), _from, _to))
}

func (r *queryResolver) EventsFromContractByTimeRange(ctx context.Context, contract string, from string, to string) ([]*model.Event, error) {
	if !(strings.HasPrefix(contract, "0x") && len(contract) == 42) {
		return nil, errors.New("Bad Contract Address")
	}

	_from, _to, err := rangeChecker(from, to, 60)
	if err != nil {
		return nil, errors.New("Bad Block Timestamp Range")
	}

	return getGraphQLCompatibleEvents(_db.GetEventsFromContractByBlockTimeRange(db, common.HexToAddress(contract), _from, _to))
}

func (r *queryResolver) EventsByBlockHash(ctx context.Context, hash string) ([]*model.Event, error) {
	if !(strings.HasPrefix(hash, "0x") && len(hash) == 66) {
		return nil, errors.New("Bad Block Hash")
	}

	return getGraphQLCompatibleEvents(_db.GetEventsByBlockHash(db, common.HexToHash(hash)))
}

func (r *queryResolver) EventsByTxHash(ctx context.Context, hash string) ([]*model.Event, error) {
	if !(strings.HasPrefix(hash, "0x") && len(hash) == 66) {
		return nil, errors.New("Bad Transaction Hash")
	}

	return getGraphQLCompatibleEvents(_db.GetEventsByTransactionHash(db, common.HexToHash(hash)))
}

func (r *queryResolver) LastXEventsFromContract(ctx context.Context, contract string, x int) ([]*model.Event, error) {
	if !(strings.HasPrefix(contract, "0x") && len(contract) == 42) {
		return nil, errors.New("Bad Contract Address")
	}

	if !(x <= 50) {
		return nil, errors.New("Too Many Events Requested")
	}

	return getGraphQLCompatibleEvents(_db.GetLastXEventsFromContract(db, common.HexToAddress(contract), x))
}

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }
