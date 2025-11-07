package database

import (
	"github.com/cockroachdb/pebble"
	"os"
	"tokenscope/common/logger"
)

/*
   Created by zyx
   Date Time: 2025/9/25
   File: tx_log_database.go
*/

var txLogDb *Database

type TxLogDatabase struct {
	database *pebble.DB
}

type ITxLogDatabase interface {
	GetDB() *pebble.DB
	Close()
}

func TxLogService() ITxLogDatabase {
	if txLogDb == nil {
		connectTxLogDB()
		return txLogDb
	}
	return txLogDb
}

func (d *Database) GetTxLogDB() *pebble.DB {
	return d.database
}

func (d *Database) CloseTxLogDb() {
	err := d.database.Close()
	if err != nil {
		logger.Logger().Errorf("Error closing database %v", err)
		return
	}
}

func connectTxLogDB() {
	ret, err := pebble.Open("./tx_logs", &pebble.Options{
		Logger: logger.Logger(),
	})
	if err != nil {
		logger.Logger().Errorf("Error opening database %v", err)
		os.Exit(0)
		return
	}
	txLogDb = &Database{
		database: ret,
	}
}
