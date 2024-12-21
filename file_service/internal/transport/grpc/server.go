package grpc

import (
	"context"
	"fmt"
	"net"

	"gitlab.crja72.ru/gospec/go19/messanger/file_service/pkg/api/crypto"
	"gitlab.crja72.ru/gospec/go19/messanger/file_service/pkg/api/file"
	"gitlab.crja72.ru/gospec/go19/messanger/file_service/pkg/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Server struct {
	grpcServer *grpc.Server
	listener   net.Listener
}

// Конструктор сервера
func New(ctx context.Context, port, cryptoPort int, service Service) (*Server, error) {
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

	conn, err := grpc.Dial(fmt.Sprintf("crypto_service:%d", cryptoPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	cryptoClient := crypto.NewCryptoServiceClient(conn)

	file.RegisterFilesServiceServer(grpcServer, NewFileService(service, logger.GetLoggerFromCtx(ctx), cryptoClient))

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
