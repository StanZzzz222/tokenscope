package models

/*
   Created by zyx
   Date Time: 2025/9/17
   File: tx.go
*/

type Tx struct {
	Hash      string `json:"hash"`
	From      string `json:"from"`
	To        string `json:"to"`
	Value     string `json:"value"`
	Data      []byte `json:"data"`
	Timestamp uint64 `json:"timestamp"`
}

type TokenTx struct {
	To    string `json:"to"`    // To address
	Value string `json:"value"` // Value: ERC20 is balance ERC721 is tokenId
}

type TxReceipt struct {
	TxHash            string `json:"tx_hash"`
	BlockHash         string `json:"block_hash"`
	BlockNumber       uint64 `json:"block_number"`
	TransactionIndex  uint64 `json:"transaction_index"`
	Status            uint64 `json:"status"`
	ConfirmCount      uint64 `json:"confirm_count"`
	GasUsed           uint64 `json:"gas_used"`
	EffectiveGasPrice string `json:"effective_gas_price"`
	CumulativeGasUsed uint64 `json:"cumulative_gas_used"`
	ContractAddress   string `json:"contract_address"`
}

func NewTx(hash, from, to, value string, data []byte, timestamp uint64) *Tx {
	return &Tx{
		Hash:      hash,
		From:      from,
		To:        to,
		Value:     value,
		Data:      data,
		Timestamp: timestamp,
	}
}
