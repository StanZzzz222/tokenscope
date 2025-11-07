package dto

import "tokenscope/models"

/*
   Created by zyx
   Date Time: 2025/9/18
   File: blockchain.go
*/

type Blockchain struct {
	BlockCount uint64           `json:"block_count"`
	SyncInfo   *models.SyncInfo `json:"sync_info"`
	Percent    float64          `json:"percent"`
}

func NewBlockchainDTO(blockCount uint64, syncInfo *models.SyncInfo, percent float64) *Blockchain {
	return &Blockchain{
		BlockCount: blockCount,
		SyncInfo:   syncInfo,
		Percent:    percent,
	}
}
