package db

import (
	"fmt"
	"log"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/itzmeanjan/ette/app/data"
	"gorm.io/gorm"
)

// GetAllBlockNumbersInRange - Returns all block numbers in given range, both inclusive
func GetAllBlockNumbersInRange(db *gorm.DB, from uint64, to uint64) []uint64 {

	var blocks []uint64

	if from < to {
		if err := db.Model(&Blocks{}).Where("number >= ? and number <= ?", from, to).Order("number asc").Select("number").Find(&blocks).Error; err != nil {

			log.Printf("[!] Failed to fetch block numbers by range : %s\n", err.Error())
			return nil

		}
	} else {
		if err := db.Model(&Blocks{}).Where("number >= ? and number <= ?", to, from).Order("number asc").Select("number").Find(&blocks).Error; err != nil {

			log.Printf("[!] Failed to fetch block numbers by range : %s\n", err.Error())
			return nil

		}
	}

	return blocks

}

// GetCurrentOldestBlockNumber - Fetches what's lowest block number present in database,
// which denotes if it's not 0, from here we can start syncing again, until we reach 0
func GetCurrentOldestBlockNumber(db *gorm.DB) uint64 {
	var number uint64

	if err := db.
		Select("min(number)").
		Table("blocks").
		Scan(&number).
		Error; err != nil {
		return 0
	}

	return number
}

// GetCurrentBlockNumber - Returns highest block number, which got processed
// by `ette`
func GetCurrentBlockNumber(db *gorm.DB) uint64 {
	var number uint64

	if err := db.Select("max(number)").
		Table("blocks").
		Scan(&number).
		Error; err != nil {
		return 0
	}

	return number
}

// GetBlockCount - Returns how many blocks currently present in database
//
// Caution : As we're dealing with very large tables
// ( with row count  ~ 10M & increasing 1 row every 2 seconds )
// this function needs to be least frequently, otherwise due to full table
// scan it'll cost us a lot
//
// Currently only using during application start up
//
// All other block count calculation requirements can be fulfilled by
// using in-memory program state holder
func GetBlockCount(db *gorm.DB) uint64 {
	var number int64

	if err := db.Model(&Blocks{}).Count(&number).Error; err != nil {
		return 0
	}

	return uint64(number)
}

// GetBlockByHash - Given blockhash finds out block related information
//
// If not found, returns nil
func GetBlockByHash(db *gorm.DB, hash common.Hash) *data.Block {
	var block data.Block

	if res := db.Model(&Blocks{}).Where("hash = ?", hash.Hex()).First(&block); res.Error != nil {
		return nil
	}

	return &block
}

// GetBlockByNumber - Fetch block using block number
//
// If not found, returns nil
func GetBlockByNumber(db *gorm.DB, number uint64) *data.Block {
	var block data.Block

	if res := db.Model(&Blocks{}).Where("number = ?", number).First(&block); res.Error != nil {
		return nil
	}

	return &block
}

// GetBlocksByNumberRange - Given block numbers as range, it'll extract out those blocks
// by number, while returning them in ascendically sorted form in terms of block numbers
//
// Note : Can return at max 10 blocks in a single query
//
// If more blocks are requested, simply to be rejected
// In that case, consider splitting them such that they satisfy criteria
func GetBlocksByNumberRange(db *gorm.DB, from uint64, to uint64) *data.Blocks {
	var blocks []*data.Block

	if res := db.Model(&Blocks{}).Where("number >= ? and number <= ?", from, to).Order("number asc").Find(&blocks); res.Error != nil {
		return nil
	}

	return &data.Blocks{
		Blocks: blocks,
	}
}

// GetBlocksByTimeRange - Given time range ( of 60 sec span at max ), returns blocks
// mined in that time span
//
// If asked to find out blocks in time span larger than 60 sec, simply drops query request
func GetBlocksByTimeRange(db *gorm.DB, from uint64, to uint64) *data.Blocks {
	var blocks []*data.Block

	if res := db.Model(&Blocks{}).Where("time >= ? and time <= ?", from, to).Order("number asc").Find(&blocks); res.Error != nil {
		return nil
	}

	return &data.Blocks{
		Blocks: blocks,
	}
}

// GetTransactionCountByBlockHash - Given block hash, finds out how many
// transactions are packed in that block
func GetTransactionCountByBlockHash(db *gorm.DB, hash common.Hash) int64 {

	var count int64

	if err := db.Model(&Transactions{}).Where("blockhash = ?", hash.Hex()).Count(&count).Error; err != nil {
		return 0
	}

	return count

}

// GetTransactionsByBlockHash - Given block hash, returns all transactions
// present in that block
func GetTransactionsByBlockHash(db *gorm.DB, hash common.Hash) *data.Transactions {
	var tx []*data.Transaction

	if res := db.Model(&Transactions{}).Where("blockhash = ?", hash.Hex()).Find(&tx); res.Error != nil {
		return nil
	}

	return &data.Transactions{
		Transactions: tx,
	}
}

// GetTransactionCountByBlockNumber - Given block number, finds out how many
// transactions are packed in that block
func GetTransactionCountByBlockNumber(db *gorm.DB, number uint64) int64 {

	var count int64

	if err := db.Model(&Transactions{}).Where("blockhash = (?)", db.Model(&Blocks{}).Where("number = ?", number).Select("hash")).Count(&count).Error; err != nil {
		return 0
	}

	return count

}

// GetTransactionsByBlockNumber - Given block number, returns all transactions
// present in that block
func GetTransactionsByBlockNumber(db *gorm.DB, number uint64) *data.Transactions {
	var tx []*data.Transaction

	if res := db.Model(&Transactions{}).Where("blockhash = (?)", db.Model(&Blocks{}).Where("number = ?", number).Select("hash")).Find(&tx); res.Error != nil {
		return nil
	}

	return &data.Transactions{
		Transactions: tx,
	}
}

// GetTransactionByHash - Given tx hash, extracts out transaction related data
func GetTransactionByHash(db *gorm.DB, hash common.Hash) *data.Transaction {
	var tx data.Transaction

	if err := db.Model(&Transactions{}).Where("hash = ?", hash.Hex()).First(&tx).Error; err != nil {
		return nil
	}

	return &tx
}

// GetTransactionCountFromAccountByBlockNumberRange - Given account address & block number range, it can find out
// how many tx(s) were sent from this account in specified block range
func GetTransactionCountFromAccountByBlockNumberRange(db *gorm.DB, account common.Address, from uint64, to uint64) int64 {

	var count int64

	if err := db.Model(&Transactions{}).Joins("left join blocks on transactions.blockhash = blocks.hash").Where("transactions.from = ? and blocks.number >= ? and blocks.number <= ?", account.Hex(), from, to).Count(&count).Error; err != nil {
		return 0
	}

	return count

}

// GetTransactionsFromAccountByBlockNumberRange - Given account address & block number range, it can find out
// all transactions which are performed from this account
func GetTransactionsFromAccountByBlockNumberRange(db *gorm.DB, account common.Address, from uint64, to uint64) *data.Transactions {
	var tx []*data.Transaction

	if err := db.Model(&Transactions{}).Joins("left join blocks on transactions.blockhash = blocks.hash").Where("transactions.from = ? and blocks.number >= ? and blocks.number <= ?", account.Hex(), from, to).Select("transactions.hash, transactions.from, transactions.to, transactions.contract, transactions.gas, transactions.gasprice, transactions.cost, transactions.nonce, transactions.state, transactions.blockhash").Find(&tx).Error; err != nil {
		return nil
	}

	return &data.Transactions{
		Transactions: tx,
	}
}

// GetTransactionCountFromAccountByBlockTimeRange - Given account address & block mining time stamp range, it can find out
// count of all tx(s) performed by this address, with in that time span
func GetTransactionCountFromAccountByBlockTimeRange(db *gorm.DB, account common.Address, from uint64, to uint64) int64 {

	var count int64

	if err := db.Model(&Transactions{}).Joins("left join blocks on transactions.blockhash = blocks.hash").Where("transactions.from = ? and blocks.time >= ? and blocks.time <= ?", account.Hex(), from, to).Count(&count).Error; err != nil {
		return 0
	}

	return count

}

// GetTransactionsFromAccountByBlockTimeRange - Given account address & block mining time stamp range, it can find out
// all tx(s) performed from this account, with in that time span
func GetTransactionsFromAccountByBlockTimeRange(db *gorm.DB, account common.Address, from uint64, to uint64) *data.Transactions {
	var tx []*data.Transaction

	if err := db.Model(&Transactions{}).Joins("left join blocks on transactions.blockhash = blocks.hash").Where("transactions.from = ? and blocks.time >= ? and blocks.time <= ?", account.Hex(), from, to).Select("transactions.hash, transactions.from, transactions.to, transactions.contract, transactions.gas, transactions.gasprice, transactions.cost, transactions.nonce, transactions.state, transactions.blockhash").Find(&tx).Error; err != nil {
		return nil
	}

	return &data.Transactions{
		Transactions: tx,
	}
}

// GetTransactionCountToAccountByBlockNumberRange - Given account address & block number range, returns #-of transactions where
// `account` was in `to` field
func GetTransactionCountToAccountByBlockNumberRange(db *gorm.DB, account common.Address, from uint64, to uint64) int64 {

	var count int64

	if err := db.Model(&Transactions{}).Joins("left join blocks on transactions.blockhash = blocks.hash").Where("transactions.to = ? and blocks.number >= ? and blocks.number <= ?", account.Hex(), from, to).Count(&count).Error; err != nil {
		return 0
	}

	return count

}

// GetTransactionsToAccountByBlockNumberRange - Given account address & block number range, returns transactions where
// `account` was in `to` field
func GetTransactionsToAccountByBlockNumberRange(db *gorm.DB, account common.Address, from uint64, to uint64) *data.Transactions {
	var tx []*data.Transaction

	if err := db.Model(&Transactions{}).Joins("left join blocks on transactions.blockhash = blocks.hash").Where("transactions.to = ? and blocks.number >= ? and blocks.number <= ?", account.Hex(), from, to).Select("transactions.hash, transactions.from, transactions.to, transactions.contract, transactions.gas, transactions.gasprice, transactions.cost, transactions.nonce, transactions.state, transactions.blockhash").Find(&tx).Error; err != nil {
		return nil
	}

	return &data.Transactions{
		Transactions: tx,
	}
}

// GetTransactionCountToAccountByBlockTimeRange - Given account address which is present in `to` field of tx(s)
// held in blocks mined with in given time range, returns those tx count
func GetTransactionCountToAccountByBlockTimeRange(db *gorm.DB, account common.Address, from uint64, to uint64) int64 {

	var count int64

	if err := db.Model(&Transactions{}).Joins("left join blocks on transactions.blockhash = blocks.hash").Where("transactions.to = ? and blocks.time >= ? and blocks.time <= ?", account.Hex(), from, to).Count(&count).Error; err != nil {
		return 0
	}

	return count

}

// GetTransactionsToAccountByBlockTimeRange - Given account address which is present in `to` field of tx(s)
// held in blocks mined with in given time range
func GetTransactionsToAccountByBlockTimeRange(db *gorm.DB, account common.Address, from uint64, to uint64) *data.Transactions {
	var tx []*data.Transaction

	if err := db.Model(&Transactions{}).Joins("left join blocks on transactions.blockhash = blocks.hash").Where("transactions.to = ? and blocks.time >= ? and blocks.time <= ?", account.Hex(), from, to).Select("transactions.hash, transactions.from, transactions.to, transactions.contract, transactions.gas, transactions.gasprice, transactions.cost, transactions.nonce, transactions.state, transactions.blockhash").Find(&tx).Error; err != nil {
		return nil
	}

	return &data.Transactions{
		Transactions: tx,
	}
}

// GetTransactionCountBetweenAccountsByBlockNumberRange - Given from & to account addresses & block number range,
// returns #-of transactions where `from` & `to` fields are matching
func GetTransactionCountBetweenAccountsByBlockNumberRange(db *gorm.DB, fromAccount common.Address, toAccount common.Address, from uint64, to uint64) int64 {

	var count int64

	if err := db.Model(&Transactions{}).Joins("left join blocks on transactions.blockhash = blocks.hash").Where("transactions.from = ? and transactions.to = ? and blocks.number >= ? and blocks.number <= ?", fromAccount.Hex(), toAccount.Hex(), from, to).Count(&count).Error; err != nil {
		return 0
	}

	return count

}

// GetTransactionsBetweenAccountsByBlockNumberRange - Given from & to account addresses & block number range,
// returns transactions where `from` & `to` fields are matching
func GetTransactionsBetweenAccountsByBlockNumberRange(db *gorm.DB, fromAccount common.Address, toAccount common.Address, from uint64, to uint64) *data.Transactions {
	var tx []*data.Transaction

	if err := db.Model(&Transactions{}).Joins("left join blocks on transactions.blockhash = blocks.hash").Where("transactions.from = ? and transactions.to = ? and blocks.number >= ? and blocks.number <= ?", fromAccount.Hex(), toAccount.Hex(), from, to).Select("transactions.hash, transactions.from, transactions.to, transactions.contract, transactions.gas, transactions.gasprice, transactions.cost, transactions.nonce, transactions.state, transactions.blockhash").Find(&tx).Error; err != nil {
		return nil
	}

	return &data.Transactions{
		Transactions: tx,
	}
}

// GetTransactionCountBetweenAccountsByBlockTimeRange - Given from & to account addresses & block mining time range,
// returns #-of transactions where `from` & `to` fields are matching
func GetTransactionCountBetweenAccountsByBlockTimeRange(db *gorm.DB, fromAccount common.Address, toAccount common.Address, from uint64, to uint64) int64 {

	var count int64

	if err := db.Model(&Transactions{}).Joins("left join blocks on transactions.blockhash = blocks.hash").Where("transactions.from = ? and transactions.to = ? and blocks.time >= ? and blocks.time <= ?", fromAccount.Hex(), toAccount.Hex(), from, to).Count(&count).Error; err != nil {
		return 0
	}

	return count

}

// GetTransactionsBetweenAccountsByBlockTimeRange - Given from & to account addresses & block mining time range,
// returns transactions where `from` & `to` fields are matching
func GetTransactionsBetweenAccountsByBlockTimeRange(db *gorm.DB, fromAccount common.Address, toAccount common.Address, from uint64, to uint64) *data.Transactions {
	var tx []*data.Transaction

	if err := db.Model(&Transactions{}).Joins("left join blocks on transactions.blockhash = blocks.hash").Where("transactions.from = ? and transactions.to = ? and blocks.time >= ? and blocks.time <= ?", fromAccount.Hex(), toAccount.Hex(), from, to).Select("transactions.hash, transactions.from, transactions.to, transactions.contract, transactions.gas, transactions.gasprice, transactions.cost, transactions.nonce, transactions.state, transactions.blockhash").Find(&tx).Error; err != nil {
		return nil
	}

	return &data.Transactions{
		Transactions: tx,
	}
}

// GetContractCreationTransactionsFromAccountByBlockNumberRange - Fetch all contract creation tx(s) from given account
// with in specific block number range
func GetContractCreationTransactionsFromAccountByBlockNumberRange(db *gorm.DB, account common.Address, from uint64, to uint64) *data.Transactions {
	var tx []*data.Transaction

	if err := db.Model(&Transactions{}).Joins("left join blocks on transactions.blockhash = blocks.hash").Where("transactions.from = ? and transactions.contract <> '' and blocks.number >= ? and blocks.number <= ?", account.Hex(), from, to).Select("transactions.hash, transactions.from, transactions.to, transactions.contract, transactions.gas, transactions.gasprice, transactions.cost, transactions.nonce, transactions.state, transactions.blockhash").Find(&tx).Error; err != nil {
		return nil
	}

	return &data.Transactions{
		Transactions: tx,
	}
}

// GetContractCreationTransactionsFromAccountByBlockTimeRange - Fetch all contract creation tx(s) from given account
// with in specific block time span range
func GetContractCreationTransactionsFromAccountByBlockTimeRange(db *gorm.DB, account common.Address, from uint64, to uint64) *data.Transactions {
	var tx []*data.Transaction

	if err := db.Model(&Transactions{}).Joins("left join blocks on transactions.blockhash = blocks.hash").Where("transactions.from = ? and transactions.contract <> '' and blocks.time >= ? and blocks.time <= ?", account.Hex(), from, to).Select("transactions.hash, transactions.from, transactions.to, transactions.contract, transactions.gas, transactions.gasprice, transactions.cost, transactions.nonce, transactions.state, transactions.blockhash").Find(&tx).Error; err != nil {
		return nil
	}

	return &data.Transactions{
		Transactions: tx,
	}
}

// GetTransactionFromAccountWithNonce - Given tx sender address & account nonce, finds out tx, satisfying condition
func GetTransactionFromAccountWithNonce(db *gorm.DB, account common.Address, nonce uint64) *data.Transaction {
	var tx data.Transaction

	if err := db.Model(&Transactions{}).Where("transactions.from = ? and transactions.nonce = ?", account.Hex(), nonce).First(&tx).Error; err != nil {
		return nil
	}

	return &tx
}

// GetEventsFromContractByBlockNumberRange - Given block number range & contract address, extracts out all
// events emitted by this contract during block span
func GetEventsFromContractByBlockNumberRange(db *gorm.DB, contract common.Address, from uint64, to uint64) *data.Events {

	var events []*data.Event

	if err := db.Model(&Events{}).Joins("left join blocks on events.blockhash = blocks.hash").Where("events.origin = ? and blocks.number >= ? and blocks.number <= ?", contract.Hex(), from, to).Select("events.origin, events.index, events.topics, events.data, events.txhash, events.blockhash").Find(&events).Error; err != nil {
		return nil
	}

	return &data.Events{
		Events: events,
	}

}

// GetEventsFromContractByBlockTimeRange - Given block time range & contract address, extracts out all
// events emitted by this contract during time span
func GetEventsFromContractByBlockTimeRange(db *gorm.DB, contract common.Address, from uint64, to uint64) *data.Events {

	var events []*data.Event

	if err := db.Model(&Events{}).Joins("left join blocks on events.blockhash = blocks.hash").Where("events.origin = ? and blocks.time >= ? and blocks.time <= ?", contract.Hex(), from, to).Select("events.origin, events.index, events.topics, events.data, events.txhash, events.blockhash").Find(&events).Error; err != nil {
		return nil
	}

	return &data.Events{
		Events: events,
	}

}

// GetEventsByBlockHash - Given block hash retrieves all events from all tx present in that block
func GetEventsByBlockHash(db *gorm.DB, blockHash common.Hash) *data.Events {
	var events []*data.Event

	if err := db.Model(&Events{}).Where("events.blockhash = ?", blockHash.Hex()).Find(&events).Error; err != nil {
		return nil
	}

	return &data.Events{
		Events: events,
	}
}

// GetEventsByTransactionHash - Given tx hash, returns all events emitted during contract interaction ( i.e. tx execution )
func GetEventsByTransactionHash(db *gorm.DB, txHash common.Hash) *data.Events {
	var events []*data.Event

	if err := db.Model(&Events{}).Where("events.txhash = ?", txHash.Hex()).Find(&events).Error; err != nil {
		return nil
	}

	return &data.Events{
		Events: events,
	}
}

// DoesItMatch - Given one event emitted by some contract & topic
// signature against which we're attempting to match the event
// it'll return boolean depending upon whether it's satisfying
// all conditions or not
func DoesItMatch(event *data.Event, topics map[uint8]string) bool {

	for k, v := range topics {

		if !(len(event.Topics) > int(k) && event.Topics[k] == v) {
			return false
		}

	}

	return true

}

// ExtractOutOnlyMatchingEvents - Extract out only those events which are having
// full match with specified event topic signatures
func ExtractOutOnlyMatchingEvents(events []*data.Event, topics map[uint8]string) *data.Events {

	sink := make([]*data.Event, 0, len(events))

	for _, e := range events {

		if DoesItMatch(e, topics) {
			sink = append(sink, e)
		}

	}

	return &data.Events{
		Events: sink,
	}

}

// EventTopicsAsString - Given event topic signatures as map,
// returns these topics as string, seperated by `,` i.e. comma
// to be used for forming DB raw query
func EventTopicsAsString(topics map[uint8]string) string {

	_topics := make([]string, 0, len(topics))

	for _, v := range topics {

		_topics = append(_topics, v)

	}

	return strings.Join(_topics, ", ")

}

// GetEventsFromContractWithTopicsByBlockNumberRange - Given block number range, contract address & topics of event log, extracts out all
// events emitted by this contract during block span with topic signatures matching
func GetEventsFromContractWithTopicsByBlockNumberRange(db *gorm.DB, contract common.Address, from uint64, to uint64, topics map[uint8]string) *data.Events {

	var events []*data.Event

	if err := db.Select("e.origin, e.index, e.topics, e.data, e.txhash, e.blockhash").
		Table("events as e").
		Joins("left join blocks as b on e.blockhash = b.hash").
		Where("e.origin = '?'", contract.Hex()).
		Where("b.number >= ?", from).
		Where("b.number <= ?", to).
		Where(fmt.Sprintf("'{%s}' <@ e.topics", EventTopicsAsString(topics))).
		Scan(&events).Error; err != nil {
		return nil
	}

	if len(events) == 0 {
		return nil
	}

	return ExtractOutOnlyMatchingEvents(events, topics)

}

// GetEventsFromContractWithTopicsByBlockTimeRange - Given time range, contract address & topics of event log, extracts out all
// events emitted by this contract during block span with topic signatures matching
func GetEventsFromContractWithTopicsByBlockTimeRange(db *gorm.DB, contract common.Address, from uint64, to uint64, topics map[uint8]string) *data.Events {

	var events []*data.Event

	if err := db.
		Select("e.origin, e.index, e.topics, e.data, e.txhash, e.blockhash").
		Table("events as e").
		Joins("left join blocks as b on e.blockhash = b.hash").
		Where("e.origin = '?'", contract.Hex()).
		Where("b.time >= ?", from).
		Where("b.time <= ?", to).
		Where(fmt.Sprintf("'{%s}' <@ e.topics", EventTopicsAsString(topics))).
		Scan(&events).
		Error; err != nil {
		return nil
	}

	if len(events) == 0 {
		return nil
	}

	return ExtractOutOnlyMatchingEvents(events, topics)

}

// GetLastXEventsFromContract - Finds out last `x` events emitted by contract
func GetLastXEventsFromContract(db *gorm.DB, contract common.Address, x int) *data.Events {

	var events []*data.Event

	if err := db.Select("e.origin, e.index, e.topics, e.data, e.txhash, e.blockhash").
		Table("events as e").
		Joins("left join blocks as b on e.blockhash = b.hash").
		Where("e.origin = '?'", contract.Hex()).
		Order("b.number desc").
		Limit(x); err != nil {
		return nil
	}

	if len(events) == 0 {
		return nil
	}

	return &data.Events{
		Events: events,
	}

}

// GetEventByBlockHashAndLogIndex - Given block hash and log index in block
// return respective event log, if any exists
func GetEventByBlockHashAndLogIndex(db *gorm.DB, hash common.Hash, index uint) *data.Event {

	var event data.Event

	if err := db.Model(&Events{}).Where("blockhash = ? and index = ?", hash.Hex(), index).First(&event).Error; err != nil {
		return nil
	}

	return &event

}

// GetEventByBlockNumberAndLogIndex - Given block number and log index in block
// return respective event log, if any exists
func GetEventByBlockNumberAndLogIndex(db *gorm.DB, number uint64, index uint) *data.Event {

	block := GetBlockByNumber(db, number)
	// seems bad block number or may be `ette`
	// hasn't synced upto this point or missed
	// this block some how
	if block == nil {
		return nil
	}

	var event data.Event

	if err := db.Model(&Events{}).Where("blockhash = ? and index = ?", block.Hash, index).First(&event).Error; err != nil {
		return nil
	}

	return &event

}
