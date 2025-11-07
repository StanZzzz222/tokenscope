package common

import (
	"encoding/hex"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
	"tokenscope/models"
)

/*
   Created by zyx
   Date Time: 2025/9/26
   File: convert.go
*/

func Convert2TokenScopeBlock(block *types.Block) *models.Block {
	var baseFee uint64 = 0
	txs := make([]*models.Tx, 0)
	ch := make(chan *models.Tx, 128)
	go func() {
		defer close(ch)
		for _, tx := range block.Transactions() {
			if tx.To() != nil {
				abiData := make([]byte, 0)
				fromAddress := getFrom(tx)
				inputHex := "0x" + hex.EncodeToString(tx.Data())
				if len(inputHex) > 0 && inputHex != "0x" {
					abiData = []byte(inputHex)
				}
				t := models.NewTx(tx.Hash().Hex(), fromAddress.Hex(), tx.To().Hex(), tx.Value().String(), abiData, block.Time())
				ch <- t
			}
		}
	}()
	for tx := range ch {
		txs = append(txs, tx)
	}
	// BaseFee only exists after the EIP-1559 upgrade
	if block.BaseFee() != nil {
		baseFee = block.BaseFee().Uint64()
	}
	return models.NewBlock(block.Hash().Hex(), block.ParentHash().Hex(), block.UncleHash().Hex(), block.TxHash().Hex(), block.Coinbase().Hex(), block.NumberU64(), block.Time(), baseFee, uint64(block.Transactions().Len()), txs)
}

// getFrom Get Transaction from address
func getFrom(tx *types.Transaction) common.Address {
	var signer types.Signer
	chainID := tx.ChainId()
	if chainID != nil && chainID.Cmp(big.NewInt(0)) != 0 {
		signer = types.LatestSignerForChainID(tx.ChainId())
	} else {
		signer = types.HomesteadSigner{}
	}
	from, err := types.Sender(signer, tx)
	if err != nil {
		from = common.HexToAddress("0x0000000000000000000000000000000000000000\n")
	}
	return from
}
