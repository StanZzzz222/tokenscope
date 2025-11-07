package handler

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"tokenscope/api/dto"
	"tokenscope/models"
	"tokenscope/repo"
)

/*
   Created by zyx
   Date Time: 2025/9/18
   File: blockchain.go
*/

func InitBlockchainRouters(r *gin.RouterGroup) {
	asset := r.Group("/blockchain")
	{
		asset.GET("/info", getInfoHandler)
	}
}

func getInfoHandler(ctx *gin.Context) {
	repository := repo.BlockchainRepository()
	blockCount := repository.GetBlockCount()
	syncInfo := repository.GetSyncInfo()
	if syncInfo == nil {
		models.Response(ctx).Error(http.StatusInternalServerError, errors.New("sync info is nil"))
		return
	}
	percent := float64(syncInfo.CurrentBlockNumber) / float64(syncInfo.LastBlockNumber)
	ret := dto.NewBlockchainDTO(blockCount, syncInfo, percent)
	models.Response(ctx).Success(http.StatusOK, ret)
}
