package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"
	"tokenscope/api"
	"tokenscope/common"
	"tokenscope/common/assets"
	"tokenscope/common/blockchain"
	"tokenscope/common/database"
	"tokenscope/common/logger"
	"tokenscope/rpc"
)

/*
   Created by zyx
   Date Time: 2025/9/11
   File: main.go
*/

func main() {
	logger.InitLogger()
	common.InitConfig()
	logger.Logger().Print("")
	logger.Logger().Print("=======================================")
	logger.Logger().Infof(" TokenScope")
	logger.Logger().Infof(" Copyright (C) 2025 StanZzz")
	logger.Logger().Print("=======================================")
	logger.Logger().Print("")
	logger.TimeTrack("Preload asset", assets.PreloadAssets)
	logger.TimeTrack("Blockchain cache", blockchain.Cache)
	logger.TimeTrack("Blockchain sync service", blockchain.BLCSync)
	logger.TimeTrack("API Service", api.Run)
	logger.TimeTrack("gRPC Service", rpc.Run)
	// Exit signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	start := time.Now()
	blockchain.BLCExitSync()
	database.BlockchainService().Close()
	logger.Logger().Infof("TokenScope shutdown took: %v ms", time.Since(start).Milliseconds())
	os.Exit(0)
}
