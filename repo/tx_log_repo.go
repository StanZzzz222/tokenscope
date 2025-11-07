package repo

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cockroachdb/pebble"
	"tokenscope/common"
	"tokenscope/common/cache"
	"tokenscope/common/database"
	"tokenscope/common/logger"
	"tokenscope/models"
)

/*
   Created by zyx
   Date Time: 2025/9/17
   File: repo.go
*/

type TxLogRepo struct{}

type ITxLogRepo interface {
	StreamTxLogs() <-chan *models.TxLog
	GetTxLogByBlockNumber(blockNumber uint64) *models.TxLog
	GetAddressTxLogsIndexed(addresses []string, ch <-chan *models.TxLog) map[int]int
	GetIndexTxLogs() []int
	HasTxLogIndexed(blockNumber uint64) bool
	StoreTxLogIndexed(blockNumber uint64)
	StoreTxLog(txLogs *models.TxLog)
}

func TxLogRepository() ITxLogRepo {
	return &TxLogRepo{}
}

func (r *TxLogRepo) StreamTxLogs() <-chan *models.TxLog {
	c := cache.Service()
	ch := make(chan *models.TxLog, 1024)
	go func() {
		defer close(ch)
		if c.Has("tx_logs_indexed") {
			wg := common.Wg(16)
			indexs := c.Get("tx_logs_indexed").([]int)
			for _, blockNumber := range indexs {
				wg.Go(func() {
					ch <- r.GetTxLogByBlockNumber(uint64(blockNumber))
				})
			}
			wg.Wait()
			return
		}
		db := database.TxLogService().GetDB()
		iter, err := db.NewIter(nil)
		if err != nil {
			logger.Logger().Errorf("New TxLog Iter error: %s", err.Error())
			return
		}
		defer func() {
			err = iter.Close()
			if err != nil {
				logger.Logger().Errorf("TxLog Iter close error: %s", err.Error())
				return
			}
		}()
		data := make([]byte, 0)
		for iter.SeekGE([]byte("tx_log_")); iter.Valid(); iter.Next() {
			data, err = iter.ValueAndErr()
			if err != nil {
				logger.Logger().Errorf("TxLog Iter get value error: %s", err.Error())
				continue
			}
			k := string(iter.Key())
			if len(k) < 7 || k[:7] != "tx_log_" {
				break
			}
			ch <- models.UnmarshalTxLog(data)
		}
	}()
	return ch
}

func (r *TxLogRepo) GetTxLogByBlockNumber(blockNumber uint64) *models.TxLog {
	data := r.streamGetTxLogByBlockNumber(blockNumber)
	return models.UnmarshalTxLog(data)
}

func (r *TxLogRepo) GetIndexTxLogs() []int {
	var ret []int
	var c = cache.Service()
	if c.Has("tx_logs_indexed") {
		ret = c.Get("tx_logs_indexed").([]int)
		return ret
	}
	db := database.TxLogService().GetDB()
	data, closer, err := db.Get([]byte("tx_logs_indexed"))
	if errors.Is(err, pebble.ErrNotFound) {
		return nil
	}
	if err != nil {
		logger.Logger().Errorf("Get tx log indexs error: %s", err.Error())
		return nil
	}
	defer func() {
		err = closer.Close()
		if err != nil {
			logger.Logger().Errorf("Closer close error: %s", err.Error())
			return
		}
	}()
	err = json.Unmarshal(data, &ret)
	if err != nil {
		logger.Logger().Errorf("IndexData un marshal error: %s", err.Error())
		return nil
	}
	return ret
}

func (r *TxLogRepo) GetAddressTxLogsIndexed(addresses []string, ch <-chan *models.TxLog) map[int]int {
	ret := make(map[int]int)
	for txLog := range ch {
		for _, address := range addresses {
			if txLog != nil && txLog.HasBloom(address) {
				ret[int(txLog.BlockNumber)] = int(txLog.BlockNumber)
				continue
			}
		}
	}
	return ret
}

func (r *TxLogRepo) HasTxLogIndexed(blockNumber uint64) bool {
	ret := false
	txLogIndexs := r.GetIndexTxLogs()
	for _, target := range txLogIndexs {
		if uint64(target) == blockNumber {
			ret = true
			break
		}
	}
	return ret
}

func (r *TxLogRepo) StoreTxLogIndexed(blockNumber uint64) {
	c := cache.Service()
	db := database.TxLogService().GetDB()
	if r.HasTxLogIndexed(blockNumber) {
		return
	}
	txLogIndexs := r.GetIndexTxLogs()
	if txLogIndexs == nil {
		txLogIndexs = make([]int, 0)
	}
	txLogIndexs = append(txLogIndexs, int(blockNumber))
	c.Set("tx_logs_indexed", txLogIndexs)
	data, err := json.Marshal(&txLogIndexs)
	if err != nil {
		logger.Logger().Errorf("Store tx log indexed error: %s", err.Error())
		return
	}
	err = db.Set([]byte("tx_logs_indexed"), data, pebble.Sync)
	if errors.Is(err, pebble.ErrClosed) {
		return
	}
	if err != nil {
		logger.Logger().Errorf("Store tx log indexed error: %s", err.Error())
		return
	}
}

func (r *TxLogRepo) StoreTxLog(txLog *models.TxLog) {
	db := database.TxLogService().GetDB()
	err := db.Set([]byte(fmt.Sprintf("tx_log_%d", txLog.BlockNumber)), txLog.Marshal(), pebble.Sync)
	if errors.Is(err, pebble.ErrClosed) {
		return
	}
	if err != nil {
		logger.Logger().Errorf("Store tx log error: %s", err.Error())
		return
	}
}

func (r *TxLogRepo) streamGetTxLogByBlockNumber(blockNumber uint64) []byte {
	db := database.TxLogService().GetDB()
	data, closer, err := db.Get([]byte(fmt.Sprintf("tx_log_%d", blockNumber)))
	if errors.Is(err, pebble.ErrNotFound) {
		return nil
	}
	if err != nil {
		logger.Logger().Errorf("Get tx log by number error: %s", err.Error())
		return nil
	}
	defer func() {
		err = closer.Close()
		if err != nil {
			logger.Logger().Errorf("Closer close error: %s", err.Error())
			return
		}
	}()
	cp := make([]byte, len(data))
	copy(cp, data)
	return cp
}
