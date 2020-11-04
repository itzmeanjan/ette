package db

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/itzmeanjan/ette/app/data"
	"gorm.io/gorm"
)

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
	var blocks []data.Block

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
	var blocks []data.Block

	if res := db.Model(&Blocks{}).Where("time >= ? and time <= ?", from, to).Order("number asc").Find(&blocks); res.Error != nil {
		return nil
	}

	return &data.Blocks{
		Blocks: blocks,
	}
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

// GetEventsFromContractWithTopic0ByBlockNumberRange - Given block number range, contract address & topic 0 of event log, extracts out all
// events emitted by this contract during block span with topic 0 signature
func GetEventsFromContractWithTopic0ByBlockNumberRange(db *gorm.DB, contract common.Address, topic common.Hash, from uint64, to uint64) *data.Events {

	var events []*data.Event

	if err := db.Model(&Events{}).Joins("left join blocks on events.blockhash = blocks.hash").Where("events.origin = ? and events.topics[1] = ? and and blocks.number >= ? and blocks.number <= ?", contract.Hex(), topic.Hex(), from, to).Select("events.origin, events.index, events.topics, events.data, events.txhash, events.blockhash").Find(&events).Error; err != nil {
		return nil
	}

	return &data.Events{
		Events: events,
	}

}

// GetEventsFromContractWithTopic0ByBlockTimeRange - Given block time range, contract address & topic 0 of event log, extracts out all
// events emitted by this contract during time span with topic 0 signature
func GetEventsFromContractWithTopic0ByBlockTimeRange(db *gorm.DB, contract common.Address, topic common.Hash, from uint64, to uint64) *data.Events {

	var events []*data.Event

	if err := db.Model(&Events{}).Joins("left join blocks on events.blockhash = blocks.hash").Where("events.origin = ? and events.topics[1] = ? and and blocks.time >= ? and blocks.time <= ?", contract.Hex(), topic.Hex(), from, to).Select("events.origin, events.index, events.topics, events.data, events.txhash, events.blockhash").Find(&events).Error; err != nil {
		return nil
	}

	return &data.Events{
		Events: events,
	}

}
