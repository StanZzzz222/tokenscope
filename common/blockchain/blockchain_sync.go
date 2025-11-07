package blockchain

import (
	"fmt"
	eCommon "github.com/ethereum/go-ethereum/common"
	eTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/puzpuzpuz/xsync/v4"
	"github.com/spf13/viper"
	"math"
	"strings"
	"sync/atomic"
	"time"
	"tokenscope/common"
	"tokenscope/common/client"
	"tokenscope/common/logger"
	"tokenscope/models"
	"tokenscope/repo"
)

/*
   Created by zyx
   Date Time: 2025/9/17
   File: blockchain_sync.go
*/

var (
	oldBlockSyncState = &atomic.Bool{}
	syncWg            = common.Wg(1)
	quitSyncSignal    = make(chan bool)
	quitSignal        = make(chan bool)
)

// BLCExitSync Blockchain exit sync
func BLCExitSync() <-chan bool {
	quitSyncSignal <- true
	quitSignal <- true
	return quitSignal
}

// BLCSync Init  blockchain sync info
func BLCSync() {
	if viper.GetBool("rpc.sync") {
		cli := client.BlockchainEVMClient()
		blsRepo := repo.BlockchainRepository()
		oldBlockSyncState.Store(true)
		if !blsRepo.HasSyncInfo() {
			start := viper.GetUint64("rpc.start")
			blockNumber := cli.GetBlockNumber()
			syncInfo := models.NewSyncInfo(start, blockNumber)
			blsRepo.SetSyncInfo(syncInfo)
		}
		syncInfo := blsRepo.GetSyncInfo()
		current := syncInfo.CurrentBlockNumber
		cli.BlockListen(func(block *eTypes.Block) {
			syncInfo = blsRepo.GetSyncInfo()
			if syncInfo != nil {
				b := common.Convert2TokenScopeBlock(block)
				current = syncInfo.CurrentBlockNumber
				syncInfo.LastBlockNumber = cli.GetBlockNumber()
				if !oldBlockSyncState.Load() {
					syncInfo.CurrentBlockNumber = cli.GetBlockNumber()
				}
				syncWg.Go(func() {
					txLogSync(cli, b)
				})
				createBlockIndexed(b)
				blsRepo.SetSyncInfo(syncInfo)
				blsRepo.StoreBlock(b)
				logger.Logger().Infof("Synced block %d, Block count: %d, Progress: %s", syncInfo.LastBlockNumber, blsRepo.GetBlockCount(), getProgress(current, syncInfo.LastBlockNumber))
			}
		}, quitSignal)
		go oldBlockSync(current, cli, blsRepo)
	} else {
		go func() {
			for {
				select {
				case <-quitSignal:
					logger.Logger().Infof("Stopping blockchain synchronization")
					select {
					case quitSignal <- true:
					default:
					}
					break
				}
			}
		}()
	}
}

// oldBlockSync Old block sync
func oldBlockSync(start uint64, cli client.IEvmClient, blsRepo repo.IRepo) {
	current := &atomic.Uint64{}
	sem := viper.GetInt("rpc.sync-sem")
	done := make(chan bool, 1)
	ch := make(chan *eTypes.Block, sem)
	blocks := xsync.NewMap[uint64, *eTypes.Block]()
	current.Store(start)
	logger.Logger().Infof("Blockchain synchronization sem: %d", sem)
	go func() {
	End:
		for {
			select {
			case <-done:
				logger.Logger().Infof("Complete blockchain synchronization")
				break End
			case <-quitSyncSignal:
				logger.Logger().Infof("Stopping blockchain synchronization")
				break End
			default:
				if block, ok := blocks.Load(start); ok {
					blockNumber := block.Number().Uint64()
					syncInfo := blsRepo.GetSyncInfo()
					if syncInfo != nil {
						b := common.Convert2TokenScopeBlock(block)
						syncInfo.CurrentBlockNumber = blockNumber
						if blockNumber > syncInfo.LastBlockNumber {
							oldBlockSyncState.Store(false)
							done <- true
							return
						}
						syncWg.Go(func() {
							txLogSync(cli, b)
						})
						createBlockIndexed(b)
						blsRepo.SetSyncInfo(syncInfo)
						blsRepo.StoreBlock(b)
						logger.Logger().Infof("Synced block %d, Block count: %d, Progress: %s", blockNumber, blsRepo.GetBlockCount(), getProgress(blockNumber, syncInfo.LastBlockNumber))
					}
					blocks.Delete(start)
					start++
				}
			}
			time.Sleep(time.Millisecond * 10)
		}
	}()
End:
	for {
		select {
		case <-quitSignal:
			quitSignal <- true
			break End
		default:
			wg := common.Wg(sem)
			for i := 0; i < sem; i++ {
				wg.Go(func() {
					blockNumber := current.Add(1) - 1
					block := cli.GetBlockByNumber(blockNumber)
					ch <- block
				})
			}
			wg.Wait()
			for i := 0; i < sem; i++ {
				select {
				case block := <-ch:
					if block != nil {
						blocks.Store(block.Number().Uint64(), block)
					}
				default:
				}
			}
			syncWg.Wait()
		}
	}
}

// txLogSync TxLog sync block
func txLogSync(cli client.IEvmClient, block *models.Block) {
	repository := repo.TxLogRepository()
	filter := models.NewFilter(block.BlockNumber, block.BlockNumber, nil, nil)
	logs := cli.GetLogs(filter)
	txEvents := make([]*models.TxEvent, 0)
	for _, log := range logs {
		if len(log.Topics) != 0 {
			method := log.Topics[0]
			methodHash := fmt.Sprintf("0x%v", method[2:10])
			values := make([]string, 0, len(log.Topics)-1)
			for _, v := range log.Topics[1:] {
				hexStr := strings.ToLower(v)
				hexStr = strings.TrimPrefix(hexStr, "0x")
				if len(hexStr) != 64 {
					values = append(values, v)
					continue
				}
				if hexStr[:24] == strings.Repeat("0", 24) {
					addr := eCommon.HexToAddress(v)
					values = append(values, addr.Hex())
				} else {
					values = append(values, "0x"+hexStr)
				}
			}
			txEvents = append(txEvents, models.NewTxEvent(log.Address.Hex(), methodHash, values))
		}
	}
	ret := models.NewTxLog(block.BlockNumber, txEvents)
	repository.StoreTxLog(ret)
}

func getProgress(start, end uint64) string {
	barLen := 30
	percent := float64(start) / float64(end)
	filled := math.Floor(percent * float64(barLen))
	bar := ""
	for i := 0; i < barLen; i++ {
		if float64(i) < filled {
			bar += "■"
		} else {
			bar += " "
		}
	}
	return fmt.Sprintf("[%s] %.2f%% (%d / %d)", bar, percent*100, start, end)
}
