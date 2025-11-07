package database

import (
	"github.com/cockroachdb/pebble"
	"os"
	"tokenscope/common/logger"
)

/*
   Created by zyx
   Date Time: 2025/9/17
   File: database.go
*/

var db *Database

type Database struct {
	database *pebble.DB
}

type IDatabase interface {
	GetDB() *pebble.DB
	Close()
}

func BlockchainService() IDatabase {
	if db == nil {
		connectDB()
		return db
	}
	return db
}

func (d *Database) GetDB() *pebble.DB {
	return d.database
}

func (d *Database) Close() {
	err := d.database.Close()
	if err != nil {
		logger.Logger().Errorf("Error closing database %v", err)
		return
	}
}

func connectDB() {
	ret, err := pebble.Open("./blockchain", &pebble.Options{
		Logger: logger.Logger(),
	})
	if err != nil {
		logger.Logger().Errorf("Error opening database %v", err)
		os.Exit(0)
		return
	}
	db = &Database{
		database: ret,
	}
}
