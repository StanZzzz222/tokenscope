package blockchain

import (
	"tokenscope/common/cache"
	"tokenscope/common/logger"
	"tokenscope/repo"
)

/*
   Created by zyx
   Date Time: 2025/9/18
   File: cache.go
*/

// Cache init cache
func Cache() {
	c := cache.Service()
	blsRepo := repo.BlockchainRepository()
	txLogRepo := repo.TxLogRepository()
	initBlockIndexed(c, blsRepo, txLogRepo)
	initBlockchainCache(c, blsRepo, txLogRepo)
}

func initBlockIndexed(c cache.ICache, blsRepo repo.IRepo, txLogRepo repo.ITxLogRepo) {
	addresses := blsRepo.GetIndexAddresses()
	if addresses != nil {
		c.Set("addresses_indexed", addresses)
		// Block
		indexs := blsRepo.GetAddressBlocksIndexed(addresses, blsRepo.StreamBlocks())
		for _, index := range indexs {
			blsRepo.StoreBlockIndexed(uint64(index))
		}
		// TxLog
		indexs = txLogRepo.GetAddressTxLogsIndexed(addresses, txLogRepo.StreamTxLogs())
		for _, index := range indexs {
			txLogRepo.StoreTxLogIndexed(uint64(index))
		}
	}
}

func initBlockchainCache(c cache.ICache, blsRepo repo.IRepo, txLogRepo repo.ITxLogRepo) {
	blocksIndexed := blsRepo.GetIndexBlocks()
	txLogsIndexed := txLogRepo.GetIndexTxLogs()
	blockCount := blsRepo.GetBlockCount()
	if blocksIndexed != nil && len(blocksIndexed) > 0 {
		c.Set("blocks_indexed", blocksIndexed)
	}
	if txLogsIndexed != nil && len(txLogsIndexed) > 0 {
		c.Set("tx_logs_indexed", blocksIndexed)
	}
	c.Set("block_count", blockCount)
	logger.Logger().Infof("Blockchain indexed loaded, indexed block count: %d", len(blocksIndexed))
	logger.Logger().Infof("TxLog indexed loaded, indexed txlog count: %d", len(txLogsIndexed))
	logger.Logger().Infof("Blockchain loaded, block count: %d", blockCount)
}
