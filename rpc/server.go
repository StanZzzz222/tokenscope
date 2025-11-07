package rpc

import (
	"fmt"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
	"tokenscope/common/logger"
	"tokenscope/rpc/middleware"
	"tokenscope/rpc/pb"
	"tokenscope/rpc/service"
)

/*
   Created by zyx
   Date Time: 2025/10/20
   File: server.go
*/

func Run() {
	if viper.GetBool("grpc.enabled") {
		host := viper.GetString("grpc.host")
		port := viper.GetInt("grpc.port")
		address := fmt.Sprintf("%s:%d", host, port)
		srv := grpc.NewServer(
			grpc.ChainUnaryInterceptor(
				middleware.LoggingInterceptor,
				middleware.AuthInterceptor,
			),
		)
		listen, err := net.Listen("tcp", address)
		if err != nil {
			logger.Logger().Errorf("gRPC Server listen falied, %v", err.Error())
			return
		}
		logger.Logger().Infof("gRPC Server listening on: %s", address)
		pb.RegisterAssetServiceServer(srv, service.AssetService())
		pb.RegisterBlockchainServiceServer(srv, service.BlockchainService())
		reflection.Register(srv)
		go func() {
			err = srv.Serve(listen)
			if err != nil {
				logger.Logger().Errorf("gRPC Server Serve falied, %v", err.Error())
				return
			}
		}()
	}
}
