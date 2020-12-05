package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/itzmeanjan/ette/app/data"
	_db "github.com/itzmeanjan/ette/app/db"
	"github.com/itzmeanjan/ette/app/rest/graph/generated"
	"github.com/itzmeanjan/ette/app/rest/graph/model"
	"gorm.io/gorm"
)

var db *gorm.DB

// GetDatabaseConnection - ...
func GetDatabaseConnection(conn *gorm.DB) {
	db = conn
}

func (r *queryResolver) BlockByHash(ctx context.Context, hash string) (*model.Block, error) {
	if !(strings.HasPrefix(hash, "0x") && len(hash) == 66) {
		return nil, errors.New("Bad Block Hash")
	}

	var block data.Block

	if res := db.Model(&_db.Blocks{}).Where("hash = ?", hash).First(&block).Error; res != nil {
		return nil, res
	}

	return &model.Block{
		Hash:                block.Hash,
		Number:              fmt.Sprintf("%d", block.Number),
		Time:                fmt.Sprintf("%d", block.Time),
		ParentHash:          block.ParentHash,
		Difficulty:          block.Difficulty,
		GasUsed:             fmt.Sprintf("%d", block.GasUsed),
		GasLimit:            fmt.Sprintf("%d", block.GasLimit),
		Nonce:               fmt.Sprintf("%d", block.Nonce),
		Miner:               block.Miner,
		Size:                block.Size,
		TransactionRootHash: block.TransactionRootHash,
		ReceiptRootHash:     block.ReceiptRootHash,
	}, nil
}

func (r *queryResolver) BlockByNumber(ctx context.Context, number string) (*model.Block, error) {
	_number, err := strconv.ParseUint(number, 10, 64)
	if err != nil {
		return nil, errors.New("Bad Block Number")
	}

	var block data.Block

	if res := db.Model(&_db.Blocks{}).Where("number = ?", _number).First(&block).Error; res != nil {
		return nil, res
	}

	return &model.Block{
		Hash:                block.Hash,
		Number:              fmt.Sprintf("%d", block.Number),
		Time:                fmt.Sprintf("%d", block.Time),
		ParentHash:          block.ParentHash,
		Difficulty:          block.Difficulty,
		GasUsed:             fmt.Sprintf("%d", block.GasUsed),
		GasLimit:            fmt.Sprintf("%d", block.GasLimit),
		Nonce:               fmt.Sprintf("%d", block.Nonce),
		Miner:               block.Miner,
		Size:                block.Size,
		TransactionRootHash: block.TransactionRootHash,
		ReceiptRootHash:     block.ReceiptRootHash,
	}, nil
}

func (r *queryResolver) Transaction(ctx context.Context, hash string) (*model.Transaction, error) {
	if !(strings.HasPrefix(hash, "0x") && len(hash) == 66) {
		return nil, errors.New("Bad Transaction Hash")
	}

	var tx data.Transaction

	if err := db.Model(&_db.Transactions{}).Where("hash = ?", hash).First(&tx).Error; err != nil {
		return nil, err
	}

	return &model.Transaction{
		Hash:      tx.Hash,
		From:      tx.From,
		To:        tx.To,
		Contract:  tx.Contract,
		Gas:       fmt.Sprintf("%d", tx.Gas),
		GasPrice:  tx.GasPrice,
		Cost:      tx.Cost,
		Nonce:     fmt.Sprintf("%d", tx.Nonce),
		State:     fmt.Sprintf("%d", tx.State),
		BlockHash: tx.BlockHash,
	}, nil
}

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }
