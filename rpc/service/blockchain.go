package service

import (
	"context"
	"errors"
	"github.com/golang/protobuf/ptypes/empty"
	"tokenscope/repo"
	"tokenscope/rpc/pb"
)

/*
   Created by zyx
   Date Time: 2025/10/24
   File: asset.go
*/

type Blockchain struct {
	pb.UnimplementedBlockchainServiceServer
}

func BlockchainService() *Blockchain {
	return &Blockchain{}
}

func (Blockchain) GetBlockchainInfo(ctx context.Context, empty *empty.Empty) (*pb.GetBlockchainInfoResponse, error) {
	repository := repo.BlockchainRepository()
	blockCount := repository.GetBlockCount()
	syncInfo := repository.GetSyncInfo()
	if syncInfo == nil {
		return nil, errors.New("sync info is nil")
	}
	percent := float64(syncInfo.CurrentBlockNumber) / float64(syncInfo.LastBlockNumber)
	return &pb.GetBlockchainInfoResponse{
		BlockchainInfo: &pb.BlockchainInfo{
			BlockCount: blockCount,
			SyncInfo: &pb.SyncInfo{
				CurrentBlockNumber: syncInfo.CurrentBlockNumber,
				LastBlockNumber:    syncInfo.LastBlockNumber,
			},
			Percent: percent,
		},
	}, nil
}
