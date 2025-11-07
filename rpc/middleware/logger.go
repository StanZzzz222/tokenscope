package middleware

import (
	"context"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"tokenscope/common/logger"
)

/*
   Created by zyx
   Date Time: 2025/10/24
   File: logger.go
*/

func LoggingInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	if viper.GetBool("grpc.log-enabled") {
		// TODO Logging
		logger.Logger().Infof("gRPC request: %v", info.FullMethod)
	}
	return handler(ctx, req)
}
