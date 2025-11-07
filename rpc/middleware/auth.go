package middleware

import (
	"context"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

/*
   Created by zyx
   Date Time: 2025/10/24
   File: auth.go
*/

func AuthInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	token := viper.GetString("grpc.token")
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok || len(md["token"]) == 0 || md["token"][0] != token {
		return nil, status.Error(codes.Unauthenticated, "Invalid auth token")
	}
	return handler(ctx, req)
}
