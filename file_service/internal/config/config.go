package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"gitlab.crja72.ru/gospec/go19/messanger/file_service/pkg/db/minio"
	"gitlab.crja72.ru/gospec/go19/messanger/file_service/pkg/db/postgres"
)

type Config struct {
	minio.MinioConfig
	postgres.PostgresConfig
	GRPCServerPort    int `env:"GRPC_SERVER_PORT" env-default:"50051"`
	CryptoServicePort int `env:"CRYPTO_SERVICE_SERVER_PORT" env-default:"50053"`
}

// Читает конфиг
func New() *Config {
	cfg := Config{}
	err := cleanenv.ReadEnv(&cfg)
	if err != nil {
		return nil
	}
	return &cfg
}
