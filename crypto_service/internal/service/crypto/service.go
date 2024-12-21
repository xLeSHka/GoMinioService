package crypto

import (
	"context"
	"fmt"
	"gitlab.crja72.ru/gospec/go19/messanger/crypto_service/pkg/crypto"
	"go.uber.org/zap"

	pb "gitlab.crja72.ru/gospec/go19/messanger/crypto_service/pkg/messenger/crypto"
)

type Service struct {
	pb.CryptoServiceServer

	secretKey []byte
	logger    *zap.Logger
}

func NewService(secretKey []byte, logger *zap.Logger) *Service {
	return &Service{secretKey: secretKey, logger: logger}
}

func (s *Service) Encrypt(_ context.Context, in *pb.Request) (*pb.Response, error) {
	if l := len(in.SecretPhrase); l != 8 {
		return nil, fmt.Errorf("incorrect length of secret phrase. expected 8, current: %d", l)
	}

	data, err := crypto.Encrypt(in.Data, append(s.secretKey, in.SecretPhrase...))
	if err != nil {
		return nil, err
	}

	return &pb.Response{Data: data}, nil
}

func (s *Service) Decrypt(_ context.Context, in *pb.Request) (*pb.Response, error) {
	if l := len(in.SecretPhrase); l != 8 {
		return nil, fmt.Errorf("incorrect length of secret phrase. expected 8, current: %d", l)
	}

	data, err := crypto.Decrypt(in.Data, append(s.secretKey, in.SecretPhrase...))
	if err != nil {
		return nil, err
	}

	return &pb.Response{Data: data}, nil
}
