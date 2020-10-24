package db

import (
	"fmt"
	"log"
	"strconv"

	"github.com/ethereum/go-ethereum/core/types"
	cfg "github.com/itzmeanjan/ette/app/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Connect - Connecting to postgresql database
func Connect() *gorm.DB {
	port, err := strconv.Atoi(cfg.Get("DB_PORT"))
	if err != nil {
		log.Fatalln("[!] ", err)
	}

	_db, err := gorm.Open(postgres.Open(fmt.Sprintf("postgresql://%s:%s@%s:%d/%s", cfg.Get("DB_USER"), cfg.Get("DB_PASSWORD"), cfg.Get("DB_HOST"), port, cfg.Get("DB_NAME"))), &gorm.Config{})
	if err != nil {
		log.Fatalln("[!] ", err)
	}

	_db.AutoMigrate(&Blocks{}, &Transactions{})
	return _db
}

// Persisting fetched block information in database
func putBlock(_db *gorm.DB, _block *types.Block) {
	if err := _db.Create(&Blocks{
		Hash:       _block.Hash().Hex(),
		Number:     _block.Number().String(),
		Time:       _block.Time(),
		ParentHash: _block.ParentHash().Hex(),
		Difficulty: _block.Difficulty().String(),
		GasUsed:    _block.GasUsed(),
		GasLimit:   _block.GasLimit(),
		Nonce:      _block.Nonce(),
	}).Error; err != nil {
		log.Println("[!] ", err)
	}
}
