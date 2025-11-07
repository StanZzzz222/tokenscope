package client

import (
	"github.com/chenzhijie/go-web3"
	"github.com/chenzhijie/go-web3/rpc"
	w3Types "github.com/chenzhijie/go-web3/types"
	"github.com/chenzhijie/go-web3/utils"
	eCommon "github.com/ethereum/go-ethereum/common"
	eTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/spf13/viper"
	"math/big"
	"reflect"
	"sync/atomic"
	"time"
	"tokenscope/common/logger"
	"tokenscope/models"
	"tokenscope/types"
	"unsafe"
)

/*
   Created by zyx
   Date Time: 2025/7/31
   File: evm_client.go
*/

var blockEvmClient IEvmClient
var contractEVMClient IEvmClient

type EvmClient struct {
	client             *web3.Web3
	chainId            uint64
	endpoint           string
	currentBlockNumber uint64
	processing         atomic.Bool
}

type IEvmClient interface {
	GetClient() *web3.Web3
	GetRPCClient() *rpc.Client
	GetChainId() uint64
	GetBlockNumber() uint64
	GetBlockByNumber(number uint64) *eTypes.Block
	GetBlockLatest() *eTypes.Block
	GetTransaction(txHash string) *eTypes.Transaction
	GetTransactionReceipt(txHash string) *eTypes.Receipt
	GetLogs(filter *models.Filter) []*w3Types.Event
	GetEstimateGas(msg *w3Types.CallMsg) uint64
	HasContractTx(txHash eCommon.Hash) bool
	BlockListen(types.BlockListenEvent, chan bool)
}

func BlockchainEVMClient() IEvmClient {
	if blockEvmClient == nil {
		blockEvmClient = newEvmClient(viper.GetString("rpc.url"))
		return blockEvmClient
	}
	return blockEvmClient
}

func ContractEVMClient() IEvmClient {
	if contractEVMClient == nil {
		contractEVMClient = newEvmClient(viper.GetString("rpc.url"))
		return contractEVMClient
	}
	return contractEVMClient
}

// newEvmClient Create a EVM Client
func newEvmClient(endpoint string) IEvmClient {
	client, err := web3.NewWeb3(endpoint)
	if err != nil {
		logger.Logger().Errorf("NewEvmClient failed, %v", err.Error())
		return nil
	}
	cli := &EvmClient{
		endpoint: endpoint,
		client:   client,
	}
	cli.chainId = cli.GetChainId()
	cli.currentBlockNumber = cli.GetBlockNumber()
	cli.resetChainID()
	return cli
}

// GetRPCClient Get rpc client
func (e *EvmClient) GetRPCClient() *rpc.Client {
	val := reflect.ValueOf(e.client).Elem().FieldByName("c")
	ptr := unsafe.Pointer(val.UnsafeAddr())
	return *(**rpc.Client)(ptr)
}

// GetClient Get evm client
func (e *EvmClient) GetClient() *web3.Web3 {
	return e.client
}

// BlockListen Block listen
func (e *EvmClient) BlockListen(callback types.BlockListenEvent, quitSignal chan bool) {
	e.chainId = e.GetChainId()
	e.currentBlockNumber = e.GetBlockNumber() - 1
	e.processBlock(callback)
	// 15s check next block
	ticker := time.NewTicker(time.Second * 15)
	go func() {
	End:
		for {
			select {
			case <-quitSignal:
				break End
			case <-ticker.C:
				e.processBlock(callback)
			}
		}
	}()
}

// GetChainId Get chainID
func (e *EvmClient) GetChainId() uint64 {
	if e.chainId <= 0 {
		chainId, err := e.client.Eth.ChainID()
		if err != nil {
			logger.Logger().Errorf("GetChainId failed: %v", err.Error())
			return 0
		}
		e.chainId = chainId.Uint64()
	}
	return e.chainId
}

// GetBlockNumber Get block number
func (e *EvmClient) GetBlockNumber() uint64 {
	if e.currentBlockNumber > 0 {
		return e.currentBlockNumber
	}
	blockNumber, err := e.client.Eth.GetBlockNumber()
	if err != nil {
		return 0
	}
	return blockNumber
}

// GetBlockLatest Get block latest
func (e *EvmClient) GetBlockLatest() *eTypes.Block {
	block, err := e.client.Eth.GetBlocByNumber(big.NewInt(int64(e.GetBlockNumber())), true)
	if err != nil {
		logger.Logger().Errorf("GetBlockLatest failed, %v", err.Error())
		time.Sleep(time.Second * 3)
		return e.GetBlockLatest()
	}
	return block
}

// GetTransaction Get transaction by txHash
func (e *EvmClient) GetTransaction(txHash string) *eTypes.Transaction {
	tx, err := e.client.Eth.GetTransactionByHash(eCommon.HexToHash(txHash))
	if err != nil {
		logger.Logger().Errorf("GetTransaction failed, %v", err.Error())
		time.Sleep(time.Second * 3)
		return e.GetTransaction(txHash)
	}
	return tx
}

// GetTransactionReceipt Get Transaction receipt by txHash
func (e *EvmClient) GetTransactionReceipt(txHash string) *eTypes.Receipt {
	tx, err := e.client.Eth.GetTransactionReceipt(eCommon.HexToHash(txHash))
	if err != nil {
		logger.Logger().Errorf("GetTransactionReceipt failed, %v", err.Error())
		time.Sleep(time.Second * 3)
		return e.GetTransactionReceipt(txHash)
	}
	return tx
}

// GetLogs Get logs
func (e *EvmClient) GetLogs(filter *models.Filter) []*w3Types.Event {
	out := make([]*w3Types.Event, 0)
	err := e.GetRPCClient().Call("eth_getLogs", &out, filter)
	if err != nil {
		logger.Logger().Errorf("GetLogs failed, %v", err.Error())
		time.Sleep(time.Second * 3)
		return e.GetLogs(filter)
	}
	return out
}

// GetEstimateGas Get estimate gas
func (e *EvmClient) GetEstimateGas(msg *w3Types.CallMsg) uint64 {
	var out string
	err := e.GetRPCClient().Call("eth_estimateGas", &out, msg)
	if err != nil {
		logger.Logger().Errorf("GetEstimateGas failed, %v", err.Error())
		time.Sleep(time.Second * 3)
		return e.GetEstimateGas(msg)
	}
	estimateGas, err := utils.ParseUint64orHex(out)
	if err != nil {
		logger.Logger().Errorf("GetEstimateGas failed, %v", err.Error())
		return 0
	}
	return estimateGas
}

// GetBlockByNumber Get block by number
func (e *EvmClient) GetBlockByNumber(number uint64) *eTypes.Block {
	block, err := e.client.Eth.GetBlocByNumber(big.NewInt(int64(number)), true)
	if err != nil {
		return nil
	}
	return block
}

// HasContractTx Checks whether the transaction is a contract transaction
func (e *EvmClient) HasContractTx(txHash eCommon.Hash) bool {
	receipt, err := e.client.Eth.GetTransactionReceipt(txHash)
	if err != nil {
		return false
	}
	return len(receipt.Logs) > 0
}

// resetChainID Reset chainID Fix: github.com/chenzhijie/go-web3 ChainID is always mainnet except testnet
func (e *EvmClient) resetChainID() {
	eth := e.GetClient().Eth
	ethVal := reflect.ValueOf(eth).Elem()
	field := ethVal.FieldByName("chainId")
	field = reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()
	field.Set(reflect.Zero(field.Type()))
}

func (e *EvmClient) processBlock(callback types.BlockListenEvent) {
	if !e.processing.CompareAndSwap(false, true) {
		return
	}
	defer e.processing.Store(false)
	nextBlockNumber := e.currentBlockNumber + 1
	if block := e.GetBlockByNumber(nextBlockNumber); block != nil {
		e.currentBlockNumber = nextBlockNumber
		callback(block)
		return
	}
}
