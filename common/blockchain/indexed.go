package blockchain

import (
	"tokenscope/common"
	"tokenscope/models"
	"tokenscope/repo"
)

/*
   Created by zyx
   Date Time: 2025/9/26
   File: indexed.go
*/

// AddressIndexed Create new address indexed
func AddressIndexed(address string, blsRepo repo.IRepo, txLogRepo repo.ITxLogRepo) {
	// Block indexed
	wg := common.Wg(2)
	wg.Go(func() {
		indexs := blsRepo.GetAddressBlocksIndexed([]string{address}, blsRepo.StreamBlocks())
		for _, index := range indexs {
			blsRepo.StoreBlockIndexed(uint64(index))
		}
	})
	// TxLog indexed
	wg.Go(func() {
		indexs := txLogRepo.GetAddressTxLogsIndexed([]string{address}, txLogRepo.StreamTxLogs())
		for _, index := range indexs {
			txLogRepo.StoreTxLogIndexed(uint64(index))
		}
	})
	wg.Wait()
}

// createBlockIndexed Create block indexed
func createBlockIndexed(block *models.Block) {
	blsRepo := repo.BlockchainRepository()
	txLogRepo := repo.TxLogRepository()
	txLog := txLogRepo.GetTxLogByBlockNumber(block.BlockNumber)
	addresses := blsRepo.GetIndexAddresses()
	for _, address := range addresses {
		if block.HasBloom(address) {
			blsRepo.StoreBlockIndexed(block.BlockNumber)
			continue
		}
		if txLog != nil && txLog.HasBloom(address) {
			txLogRepo.StoreTxLogIndexed(block.BlockNumber)
			continue
		}
	}
}
