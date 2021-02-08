package graph

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/itzmeanjan/ette/app/data"
	_db "github.com/itzmeanjan/ette/app/db"
	"github.com/itzmeanjan/ette/app/rest/graph/model"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

var db *gorm.DB

// GetDatabaseConnection - Passing already connected database handle to this package,
// so that it can be used for handling database queries for resolving graphQL queries
func GetDatabaseConnection(conn *gorm.DB) {
	db = conn
}

// Attempting to recover router context i.e. which holds client `APIKey` in request header,
// in graphql handler context, so that we can do some accounting job
func routerContextFromGraphQLContext(ctx context.Context) (*gin.Context, error) {

	ginContext := ctx.Value("RouterContextInGraphQL")
	if ginContext == nil {
		return nil, errors.New("Failed to retrieve router context")
	}

	gc, ok := ginContext.(*gin.Context)
	if !ok {
		return nil, errors.New("Type assert of router context failed")
	}

	return gc, nil

}

// Attempts to extract out `APIKey` used passed along
// with request, in graphql handler function, to be used for
// doing some book keeping work
func getAPIKey(ctx context.Context) string {

	routerCtx, err := routerContextFromGraphQLContext(ctx)
	if err != nil {

		log.Printf("[!] Failed to get `APIKey` : %s\n", err.Error())
		return ""

	}

	return routerCtx.GetHeader("APIKey")

}

// Attempts to recover `APIKey` from router context, which is
// then used for looking up user, so that data delivery information can
// be persisted into DB
func doBookKeeping(ctx context.Context, _data []byte) error {

	if _data == nil {
		return errors.New("JSON marshalling failed")
	}

	user := _db.GetUserFromAPIKey(db, getAPIKey(ctx))
	if user == nil {
		return errors.New("Failed to get user from `APIKey`")
	}

	_db.PutDataDeliveryInfo(db, user.Address, "/v1/graphql", uint64(len(_data)))
	return nil

}

// Converting block data to graphQL compatible data structure
func getGraphQLCompatibleBlock(ctx context.Context, block *data.Block, bookKeeping bool) (*model.Block, error) {

	if block == nil {
		return nil, errors.New("Found nothing")
	}

	// to be `false` when calling from `getGraphQLCompatibleBlocks(...)`
	// because that function will then take care of it's own book keeping logic
	if bookKeeping {
		if err := doBookKeeping(ctx, block.ToJSON()); err != nil {
			return nil, errors.New("Book keeping failed")
		}
	}

	extraData := ""
	if _h := hex.EncodeToString(block.ExtraData); _h != "" {
		extraData = fmt.Sprintf("0x%s", _h)
	}

	return &model.Block{
		Hash:            block.Hash,
		Number:          fmt.Sprintf("%d", block.Number),
		Time:            fmt.Sprintf("%d", block.Time),
		ParentHash:      block.ParentHash,
		Difficulty:      block.Difficulty,
		GasUsed:         fmt.Sprintf("%d", block.GasUsed),
		GasLimit:        fmt.Sprintf("%d", block.GasLimit),
		Nonce:           block.Nonce,
		Miner:           block.Miner,
		Size:            block.Size,
		StateRootHash:   block.StateRootHash,
		UncleHash:       block.UncleHash,
		TxRootHash:      block.TransactionRootHash,
		ReceiptRootHash: block.ReceiptRootHash,
		ExtraData:       extraData,
	}, nil

}

// Converting block array to graphQL compatible data structure
func getGraphQLCompatibleBlocks(ctx context.Context, blocks *data.Blocks) ([]*model.Block, error) {
	if blocks == nil {
		return nil, errors.New("Found nothing")
	}

	if !(len(blocks.Blocks) > 0) {
		return nil, errors.New("Found nothing")
	}

	if err := doBookKeeping(ctx, blocks.ToJSON()); err != nil {
		return nil, errors.New("Book keeping failed")
	}

	_blocks := make([]*model.Block, len(blocks.Blocks))

	for k, v := range blocks.Blocks {
		_v, _ := getGraphQLCompatibleBlock(ctx, v, false)
		_blocks[k] = _v
	}

	return _blocks, nil
}

// Converting transaction data to graphQL compatible data structure
func getGraphQLCompatibleTransaction(ctx context.Context, tx *data.Transaction, bookKeeping bool) (*model.Transaction, error) {
	if tx == nil {
		return nil, errors.New("Found nothing")
	}

	// to be `false` when calling from `getGraphQLCompatibleTransactions(...)`
	// because that function will then take care of it's own book keeping logic
	if bookKeeping {
		if err := doBookKeeping(ctx, tx.ToJSON()); err != nil {
			return nil, errors.New("Book keeping failed")
		}
	}

	data := ""
	if _h := hex.EncodeToString(tx.Data); _h != "" {
		data = fmt.Sprintf("0x%s", _h)
	}

	if !strings.HasPrefix(tx.Contract, "0x") {
		return &model.Transaction{
			Hash:      tx.Hash,
			From:      tx.From,
			To:        tx.To,
			Contract:  "",
			Value:     tx.Value,
			Data:      data,
			Gas:       fmt.Sprintf("%d", tx.Gas),
			GasPrice:  tx.GasPrice,
			Cost:      tx.Cost,
			Nonce:     fmt.Sprintf("%d", tx.Nonce),
			State:     fmt.Sprintf("%d", tx.State),
			BlockHash: tx.BlockHash,
		}, nil
	}

	return &model.Transaction{
		Hash:      tx.Hash,
		From:      tx.From,
		To:        "",
		Contract:  tx.Contract,
		Value:     tx.Value,
		Data:      data,
		Gas:       fmt.Sprintf("%d", tx.Gas),
		GasPrice:  tx.GasPrice,
		Cost:      tx.Cost,
		Nonce:     fmt.Sprintf("%d", tx.Nonce),
		State:     fmt.Sprintf("%d", tx.State),
		BlockHash: tx.BlockHash,
	}, nil
}

// Converting transaction array to graphQL compatible data structure
func getGraphQLCompatibleTransactions(ctx context.Context, tx *data.Transactions) ([]*model.Transaction, error) {
	if tx == nil {
		return nil, errors.New("Found nothing")
	}

	if !(len(tx.Transactions) > 0) {
		return nil, errors.New("Found nothing")
	}

	if err := doBookKeeping(ctx, tx.ToJSON()); err != nil {
		return nil, errors.New("Book keeping failed")
	}

	_tx := make([]*model.Transaction, len(tx.Transactions))

	for k, v := range tx.Transactions {
		_v, _ := getGraphQLCompatibleTransaction(ctx, v, false)
		_tx[k] = _v
	}

	return _tx, nil
}

// Converting event data to graphQL compatible data structure
func getGraphQLCompatibleEvent(ctx context.Context, event *data.Event, bookKeeping bool) (*model.Event, error) {
	if event == nil {
		return nil, errors.New("Found nothing")
	}

	// to be `false` when calling from `getGraphQLCompatibleEvents(...)`
	// because that function will then take care of it's own book keeping logic
	if bookKeeping {
		if err := doBookKeeping(ctx, event.ToJSON()); err != nil {
			return nil, errors.New("Book keeping failed")
		}
	}

	data := ""
	if _h := hex.EncodeToString(event.Data); _h != "" && _h != strings.Repeat("0", 64) {
		data = fmt.Sprintf("0x%s", _h)
	}

	return &model.Event{
		Origin:    event.Origin,
		Index:     fmt.Sprintf("%d", event.Index),
		Topics:    getTopicSignaturesAsStringSlice(event.Topics),
		Data:      data,
		TxHash:    event.TransactionHash,
		BlockHash: event.BlockHash,
	}, nil
}

// Converting event array to graphQL compatible data structure
func getGraphQLCompatibleEvents(ctx context.Context, events *data.Events) ([]*model.Event, error) {
	if events == nil {
		return nil, errors.New("Found nothing")
	}

	if !(len(events.Events) > 0) {
		return nil, errors.New("Found nothing")
	}

	if err := doBookKeeping(ctx, events.ToJSON()); err != nil {
		return nil, errors.New("Book keeping failed")
	}

	_events := make([]*model.Event, len(events.Events))

	for k, v := range events.Events {
		_v, _ := getGraphQLCompatibleEvent(ctx, v, false)
		_events[k] = _v
	}

	return _events, nil
}

func getTopicSignaturesAsStringSlice(topics pq.StringArray) []string {
	_tmp := make([]string, len(topics))

	for k, v := range topics {
		_tmp[k] = v
	}

	return _tmp
}

// FillUpTopicArray - Creates a topic signature array of length
// 4, while putting all elements passed from graphQL query & appending
// empty strings, in remaining places
func FillUpTopicArray(topics []string) []string {

	if len(topics) == 4 {
		return topics
	}

	result := make([]string, 0, 4)
	result = append(result, topics...)

	i := 0
	target := 4 - len(topics)

	for i < target {

		result = append(result, "")
		i++

	}

	return result

}
