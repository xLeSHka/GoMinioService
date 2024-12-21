package server

import (
	"gitlab.crja72.ru/gospec/go19/messanger/crypto_service/internal/service/crypto"
	pb "gitlab.crja72.ru/gospec/go19/messanger/crypto_service/pkg/messenger/crypto"
	"google.golang.org/grpc"
	"net"
)

type Server interface {
	Start() error
	Shutdown() error
}

type server struct {
	lis        net.Listener
	grpcServer *grpc.Server
}

func NewServer(grpcAddress string, serviceServer *crypto.Service) (Server, error) {
	// Создаём gRPC listener
	lis, err := net.Listen("tcp", grpcAddress)
	if err != nil {
		return nil, err
	}

	// Создаём gRPC сервер
	grpcServer := grpc.NewServer()
	pb.RegisterCryptoServiceServer(grpcServer, serviceServer)

	return &server{
		lis:        lis,
		grpcServer: grpcServer,
	}, nil
}

// Start запускает сервер. Метод является блокирующим
func (s *server) Start() error {
	return s.grpcServer.Serve(s.lis)
}

// Shutdown останавливает сервер
func (s *server) Shutdown() error {
	s.grpcServer.GracefulStop()
	return nil
}
