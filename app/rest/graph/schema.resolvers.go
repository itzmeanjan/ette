package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/itzmeanjan/ette/app/data"
	_db "github.com/itzmeanjan/ette/app/db"
	"github.com/itzmeanjan/ette/app/rest/graph/generated"
	"github.com/itzmeanjan/ette/app/rest/graph/model"
	"gorm.io/gorm"
)

var db *gorm.DB

// GetDatabaseConnection - Passing already connected database handle to this package,
// to be used for resolving graphQL queries
func GetDatabaseConnection(conn *gorm.DB) {
	db = conn
}

func (r *queryResolver) Block(ctx context.Context, hash string) (*model.Block, error) {
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

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }
