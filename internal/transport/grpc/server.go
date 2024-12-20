package grpc

import (
	"context"
	"fmt"
	"net"

	"github.com/xLeSHka/GoMinioService/pkg/api/file"
	"github.com/xLeSHka/GoMinioService/pkg/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Server struct {
	grpcServer *grpc.Server
	listener   net.Listener
}

// Конструктор сервера
func New(ctx context.Context, port int, service Service) (*Server, error) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	opts := []grpc.ServerOption{
		grpc.ChainStreamInterceptor(
			StreamLoggerInterceptor(logger.GetLoggerFromCtx(ctx)),
		),
		grpc.ChainUnaryInterceptor(
			UnaryLoggerInterceptor(logger.GetLoggerFromCtx(ctx)),
		),
	}
	grpcServer := grpc.NewServer(opts...)

	file.RegisterFilesServiceServer(grpcServer, NewFileService(service, logger.GetLoggerFromCtx(ctx)))

	return &Server{grpcServer, lis}, nil
}

// Запуск сервера
func (s *Server) Start(ctx context.Context) error {
	logger.GetLoggerFromCtx(ctx).Info(ctx, "starting grpc server", zap.Int("port", s.listener.Addr().(*net.TCPAddr).Port))
	return s.grpcServer.Serve(s.listener)
}

// плавная остановка сервера
func (s *Server) Stop(ctx context.Context) {
	s.grpcServer.GracefulStop()
	logger.GetLoggerFromCtx(ctx).Info(ctx, "grpc server stopped")
}
