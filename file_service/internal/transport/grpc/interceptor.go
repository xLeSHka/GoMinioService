package grpc

import (
	"context"

	"gitlab.crja72.ru/gospec/go19/messanger/file_service/pkg/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)
//логируем унари запрос к грпс
func UnaryLoggerInterceptor(l logger.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		l.Info(ctx, "request started", zap.String("method", info.FullMethod))
		return handler(ctx, req)
	}
}
//логируем стрим запрос к грпс
func StreamLoggerInterceptor(l logger.Logger) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		l.Info(context.Background(), "request started", zap.String("method", info.FullMethod))
		return handler(srv, ss)
	}
}
