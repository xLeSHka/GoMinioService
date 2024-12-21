package main

import (
	"fmt"
	"gitlab.crja72.ru/gospec/go19/messanger/crypto_service/internal/config"
	"gitlab.crja72.ru/gospec/go19/messanger/crypto_service/internal/server"
	"gitlab.crja72.ru/gospec/go19/messanger/crypto_service/internal/service/crypto"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
)

// Получение конфига 2 методами
func getConfig(logger *zap.Logger) config.Config {
	const configPath = "config.yml"

	cfg, err := config.ReadFromFile(configPath)
	if err != nil {
		logger.Warn("Failed to get config from file. Attempting to get from environment variables",
			zap.Error(err))

		cfg, err := config.ReadFromEnv()
		if err != nil {
			logger.Fatal("Failed to get config from environment variables. Server stop...",
				zap.Error(err))
		}
		logger.Info("It was possible to get the config from the environment variables. Continue...")
		return cfg
	}

	return cfg
}

func main() {
	// Создаём логгер
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(fmt.Sprintf("Failed to create a logger instance: %v", err))
	}
	//goland:noinspection GoUnhandledErrorResult
	defer logger.Sync()

	// Получаем конфиг
	cfg := getConfig(logger)

	// Создаём сервис
	service := crypto.NewService([]byte(cfg.SecretKey), logger)

	// Создаём сервер
	serv, err := server.NewServer(fmt.Sprintf(":%d", cfg.GRPCPort), service)
	if err != nil {
		logger.Fatal("Failed to create a server instance", zap.Error(err))
	}

	// Запускаем сервер
	go func() {
		if err := serv.Start(); err != nil {
			logger.Fatal("Server startup error", zap.Error(err))
		}
	}()

	logger.Info("Server starting...")

	// Ждём-с
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	<-c

	// Останавливаем сервер
	if err := serv.Shutdown(); err != nil {
		logger.Fatal("Server shutdown error", zap.Error(err))
	}

	logger.Info("Server stopped")
}
