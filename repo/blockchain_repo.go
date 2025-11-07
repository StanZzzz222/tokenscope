package repo

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cockroachdb/pebble"
	"strings"
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

type Repo struct{}

type IRepo interface {
	StreamBlocks() <-chan *models.Block
	GetBlockByNumber(blockNumber uint64) *models.Block
	GetAddressBlocksIndexed(addresses []string, ch <-chan *models.Block) map[int]int
	GetIndexAddresses() []string
	GetIndexBlocks() []int
	GetBlockCount() uint64
	GetTxByHash(hash string) *models.Tx
	GetSyncInfo() *models.SyncInfo
	SetSyncInfo(syncInfo *models.SyncInfo)
	StoreBlock(block *models.Block)
	StoreBlockIndexed(blockNumber uint64)
	StoreAddressIndexed(address string)
	HasBlockIndexed(blockNumber uint64) bool
	HasAddressIndexed(address string) bool
	HasSyncInfo() bool
}

func BlockchainRepository() IRepo {
	return &Repo{}
}

func (r *Repo) StreamBlocks() <-chan *models.Block {
	c := cache.Service()
	ch := make(chan *models.Block, 512)
	go func() {
		defer close(ch)
		if c.Has("blocks_indexed") {
			indexs := c.Get("blocks_indexed").([]int)
			wg := common.Wg(16)
			for _, blockNumber := range indexs {
				wg.Go(func() {
					ch <- r.GetBlockByNumber(uint64(blockNumber))
				})
			}
			wg.Wait()
			return
		}
		db := database.BlockchainService().GetDB()
		iter, err := db.NewIter(nil)
		if err != nil {
			logger.Logger().Errorf("New Block Iter error: %s", err.Error())
			return
		}
		defer func() {
			err = iter.Close()
			if err != nil {
				logger.Logger().Errorf("Block Iter close error: %s", err.Error())
				return
			}
		}()
		data := make([]byte, 0)
		for iter.SeekGE([]byte("block_")); iter.Valid(); iter.Next() {
			key := string(iter.Key())
			if len(key) < 6 || key[:6] != "block_" {
				break
			}
			data, err = iter.ValueAndErr()
			if err != nil {
				logger.Logger().Errorf("Block Iter get value error: %s", err.Error())
				continue
			}
			ch <- models.UnmarshalBlock(data)
		}
	}()
	return ch
}

func (r *Repo) GetBlockByNumber(blockNumber uint64) *models.Block {
	data := r.streamGetBlockByNumber(blockNumber)
	return models.UnmarshalBlock(data)
}

func (r *Repo) GetIndexBlocks() []int {
	var ret []int
	var c = cache.Service()
	if c.Has("blocks_indexed") {
		ret = c.Get("blocks_indexed").([]int)
		return ret
	}
	db := database.BlockchainService().GetDB()
	data, closer, err := db.Get([]byte("blocks_indexed"))
	if errors.Is(err, pebble.ErrNotFound) {
		return nil
	}
	if err != nil {
		logger.Logger().Errorf("Get block indexs error: %s", err.Error())
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
		logger.Logger().Errorf("IndexData unmarshal error: %s", err.Error())
		return nil
	}
	return ret
}

func (r *Repo) GetIndexAddresses() []string {
	var ret []string
	var c = cache.Service()
	if c.Has("addresses_indexed") {
		ret = c.Get("addresses_indexed").([]string)
		return ret
	}
	db := database.BlockchainService().GetDB()
	data, closer, err := db.Get([]byte("addresses_indexed"))
	if errors.Is(err, pebble.ErrNotFound) {
		return nil
	}
	if err != nil {
		logger.Logger().Errorf("Get address indexs error: %s", err.Error())
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

func (r *Repo) GetAddressBlocksIndexed(addresses []string, ch <-chan *models.Block) map[int]int {
	ret := make(map[int]int)
	for block := range ch {
		for _, address := range addresses {
			if block != nil && block.HasBloom(address) {
				ret[int(block.BlockNumber)] = int(block.BlockNumber)
				continue
			}
		}
	}
	return ret
}

func (r *Repo) HasAddressIndexed(address string) bool {
	ret := false
	addresses := r.GetIndexAddresses()
	for _, target := range addresses {
		if target == address {
			ret = true
			break
		}
	}
	return ret
}

func (r *Repo) HasBlockIndexed(blockNumber uint64) bool {
	ret := false
	blockIndexs := r.GetIndexBlocks()
	for _, target := range blockIndexs {
		if uint64(target) == blockNumber {
			ret = true
			break
		}
	}
	return ret
}

func (r *Repo) GetBlockCount() uint64 {
	var ret uint64
	c := cache.Service()
	if c.Has("block_count") {
		count := c.Get("block_count").(uint64)
		return count
	}
	db := database.BlockchainService().GetDB()
	iter, err := db.NewIter(nil)
	if err != nil {
		return 0
	}
	defer func() {
		err = iter.Close()
		if err != nil {
			logger.Logger().Errorf("Block Iter close error: %s", err.Error())
			return
		}
	}()
	for iter.SeekGE([]byte("block_")); iter.Valid(); iter.Next() {
		ret += 1
	}
	return ret
}

func (r *Repo) GetTxByHash(hash string) *models.Tx {
	var ret *models.Tx
	ret = iterBlockGetTx(hash, r.StreamBlocks())
	return ret
}

func (r *Repo) HasSyncInfo() bool {
	db := database.BlockchainService().GetDB()
	_, closer, err := db.Get([]byte("sync_info"))
	if errors.Is(err, pebble.ErrNotFound) {
		return false
	}
	if err != nil {
		logger.Logger().Errorf("Has sync info error: %s", err.Error())
		return false
	}
	if err = closer.Close(); err != nil {
		return false
	}
	return true
}

func (r *Repo) GetSyncInfo() *models.SyncInfo {
	db := database.BlockchainService().GetDB()
	data, closer, err := db.Get([]byte("sync_info"))
	if errors.Is(err, pebble.ErrClosed) {
		return nil
	}
	if err != nil {
		logger.Logger().Errorf("Get sync info error: %s", err.Error())
		return nil
	}
	if err = closer.Close(); err != nil {
		logger.Logger().Errorf("Close close db error: %s", err.Error())
		return nil
	}
	ret := models.UnmarshalSyncInfo(data)
	return ret
}

func (r *Repo) SetSyncInfo(syncInfo *models.SyncInfo) {
	data := syncInfo.Marshal()
	db := database.BlockchainService().GetDB()
	err := db.Set([]byte("sync_info"), data, pebble.Sync)
	if errors.Is(err, pebble.ErrClosed) {
		return
	}
	if err != nil {
		logger.Logger().Errorf("Set sync info error: %s", err.Error())
		return
	}
}

func (r *Repo) StoreBlock(block *models.Block) {
	c := cache.Service()
	db := database.BlockchainService().GetDB()
	err := db.Set([]byte(fmt.Sprintf("block_%d", block.BlockNumber)), block.Marshal(), pebble.Sync)
	if errors.Is(err, pebble.ErrClosed) {
		return
	}
	if err != nil {
		logger.Logger().Errorf("Store block error: %s", err.Error())
		return
	}
	if c.Has("block_count") {
		count := c.Get("block_count").(uint64)
		c.Set("block_count", count+1)
	}
}

func (r *Repo) StoreBlockIndexed(blockNumber uint64) {
	c := cache.Service()
	db := database.BlockchainService().GetDB()
	if r.HasBlockIndexed(blockNumber) {
		return
	}
	blockIndexs := r.GetIndexBlocks()
	if blockIndexs == nil {
		blockIndexs = make([]int, 0)
	}
	blockIndexs = append(blockIndexs, int(blockNumber))
	c.Set("blocks_indexed", blockIndexs)
	data, err := json.Marshal(&blockIndexs)
	if err != nil {
		logger.Logger().Errorf("Store block indexed error: %s", err.Error())
		return
	}
	err = db.Set([]byte("blocks_indexed"), data, pebble.Sync)
	if errors.Is(err, pebble.ErrClosed) {
		return
	}
	if err != nil {
		logger.Logger().Errorf("Store block indexed error: %s", err.Error())
		return
	}
}

func (r *Repo) StoreAddressIndexed(address string) {
	c := cache.Service()
	db := database.BlockchainService().GetDB()
	if !r.HasAddressIndexed(address) {
		addresses := r.GetIndexAddresses()
		if addresses == nil {
			addresses = make([]string, 0)
		}
		addresses = append(addresses, address)
		c.Set("addresses_indexed", addresses)
		data, err := json.Marshal(&addresses)
		if err != nil {
			logger.Logger().Errorf("Store address indexed error: %s", err.Error())
			return
		}
		err = db.Set([]byte("addresses_indexed"), data, pebble.Sync)
		if errors.Is(err, pebble.ErrClosed) {
			return
		}
		if err != nil {
			logger.Logger().Errorf("Store address indexed error: %s", err.Error())
			return
		}
	}
}

func (r *Repo) streamGetBlockByNumber(blockNumber uint64) []byte {
	db := database.BlockchainService().GetDB()
	data, closer, err := db.Get([]byte(fmt.Sprintf("block_%d", blockNumber)))
	if errors.Is(err, pebble.ErrNotFound) {
		return nil
	}
	if err != nil {
		logger.Logger().Errorf("Get block by number error: %s", err.Error())
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

func iterBlockGetTx(hash string, ch <-chan *models.Block) (ret *models.Tx) {
	wg := common.Wg(32)
	for block := range ch {
		if block.HasBloom(hash) {
			wg.Go(func() {
				for _, tx := range block.Txs {
					if strings.ToLower(tx.Hash) == strings.ToLower(hash) {
						ret = tx
						return
					}
				}
			})
		}
	}
	wg.Wait()
	return ret
}
