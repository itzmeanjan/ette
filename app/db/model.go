package db

import "github.com/lib/pq"

// Tabler - ...
type Tabler interface {
	TableName() string
}

// Blocks - Mined block info holder table model
type Blocks struct {
	Hash         string       `gorm:"column:hash;type:char(66);primaryKey"`
	Number       string       `gorm:"column:number;type:varchar;not null"`
	Time         uint64       `gorm:"column:time;type:bigint;not null"`
	ParentHash   string       `gorm:"column:parenthash;type:char(66);not null"`
	Difficulty   string       `gorm:"column:difficulty;type:varchar;not null"`
	GasUsed      uint64       `gorm:"column:gasused;type:bigint;not null"`
	GasLimit     uint64       `gorm:"column:gaslimit;type:bigint;not null"`
	Nonce        uint64       `gorm:"column:nonce;type:bigint;not null"`
	Transactions Transactions `gorm:"foreignKey:blockhash"`
	Events       Events       `gorm:"foreignKey:blockhash"`
}

// TableName - Overriding default table name
func (Blocks) TableName() string {
	return "blocks"
}

// Transactions - Blockchain transaction holder table model
type Transactions struct {
	Hash      string `gorm:"column:hash;type:char(66);primaryKey"`
	From      string `gorm:"column:from;type:char(42);not null"`
	To        string `gorm:"column:to;type:char(42); not null"`
	Gas       uint64 `gorm:"column:gas;type:bigint;not null"`
	GasPrice  string `gorm:"column:gasprice;type:varchar;not null"`
	Cost      string `gorm:"column:cost;type:varchar;not null"`
	Nonce     uint64 `gorm:"column:nonce;type:bigint;not null"`
	State     uint64 `gorm:"column:state;type:smallint;not null"`
	BlockHash string `gorm:"column:blockhash;type:char(66);not null"`
	Events    Events `gorm:"foreignKey:txhash"`
}

// TableName - Overriding default table name
func (Transactions) TableName() string {
	return "transactions"
}

// Events - Events emitted from smart contracts to be held in this table
type Events struct {
	Origin          string         `gorm:"column:origin;type:char(42);not null"`
	Index           uint           `gorm:"column:index;type:integer;not null;primaryKey"`
	Topics          pq.StringArray `gorm:"column:topics;type:text[];not null"`
	Data            []byte         `gorm:"column:data;type:bytea"`
	TransactionHash string         `gorm:"column:txhash;type:char(66);not null"`
	BlockHash       string         `gorm:"column:blockhash;type:char(66);not null;primaryKey"`
}

// TableName - Overriding default table name
func (Events) TableName() string {
	return "events"
}
