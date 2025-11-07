package models

import (
	"encoding/json"
	"tokenscope/common/logger"
)

/*
   Created by zyx
   Date Time: 2025/9/17
   File: sync.go
*/

type SyncInfo struct {
	CurrentBlockNumber uint64 `json:"current_block_number"`
	LastBlockNumber    uint64 `json:"last_block_number"`
}

func NewSyncInfo(currentBlockNumber uint64, lastBlockNumber uint64) *SyncInfo {
	return &SyncInfo{
		CurrentBlockNumber: currentBlockNumber,
		LastBlockNumber:    lastBlockNumber,
	}
}

func UnmarshalSyncInfo(data []byte) *SyncInfo {
	ret := new(SyncInfo)
	err := json.Unmarshal(data, &ret)
	if err != nil {
		logger.Logger().Errorf("Unmarshal SyncInfo err:%v", err)
		return nil
	}
	return ret
}

func (s *SyncInfo) Marshal() []byte {
	data, err := json.Marshal(s)
	if err != nil {
		logger.Logger().Errorf("Marshal SyncInfo err:%v", err)
		return nil
	}
	return data
}
