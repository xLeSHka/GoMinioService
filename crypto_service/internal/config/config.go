package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	SecretKey string `yaml:"secret-key" env:"SECRET_KEY" env-default:"123456789012345678901234"`
	GRPCPort  uint16 `yaml:"grpc-port" env:"GRPC_PORT" env-default:"50053"`
}

func validateConfig(config Config) error {
	if length := len(config.SecretKey); length != 24 {
		return fmt.Errorf("incorrect key length is specified: %d. acceptable: 24", length)
	}

	return nil
}

// ReadFromFile получает конфиг из файла
func ReadFromFile(path string) (Config, error) {
	var cfg Config

	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		return Config{}, err
	}

	if err := validateConfig(cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// ReadFromEnv получает конфиг из переменных среды
func ReadFromEnv() (Config, error) {
	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return Config{}, err
	}

	if err := validateConfig(cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
