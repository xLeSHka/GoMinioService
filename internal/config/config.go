package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/xLeSHka/GoMinioService/pkg/db/minio"
	"github.com/xLeSHka/GoMinioService/pkg/db/postgres"
)

type Config struct {
	minio.MinioConfig
	postgres.PostgresConfig
	GRPCServerPort int `env:"GRPC_SERVER_PORT" env-default:"50051"`
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
