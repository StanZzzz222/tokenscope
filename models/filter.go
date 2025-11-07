package models

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
)

/*
   Created by zyx
   Date Time: 2025/9/25
   File: filter.go
*/

type Filter struct {
	Address   *common.Address `json:"address,omitempty"`
	FromBlock string          `json:"fromBlock,omitempty"`
	ToBlock   string          `json:"toBlock,omitempty"`
	Topics    []string        `json:"topics,omitempty"`
}

func NewFilter(fromBlock, toBlock uint64, address *common.Address, topics []string) *Filter {
	return &Filter{
		Address:   address,
		FromBlock: fmt.Sprintf("0x%x", fromBlock),
		ToBlock:   fmt.Sprintf("0x%x", toBlock),
		Topics:    topics,
	}
}
