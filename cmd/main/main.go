package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/xLeSHka/GoMinioService/internal/config"
	"github.com/xLeSHka/GoMinioService/internal/repository"
	"github.com/xLeSHka/GoMinioService/internal/service"
	"github.com/xLeSHka/GoMinioService/internal/transport/grpc"
	"github.com/xLeSHka/GoMinioService/pkg/db/minio"
	"github.com/xLeSHka/GoMinioService/pkg/db/postgres"
	"github.com/xLeSHka/GoMinioService/pkg/logger"
	"go.uber.org/zap"
)

var (
	serviceName = "files"
)

func main() {
	// создаем логер и пихаем го в контектс
	ctx := context.Background()
	mainLogger := logger.New(serviceName)
	ctx = context.WithValue(ctx, logger.LoggerKey, mainLogger)
	//читаем конфиг
	cfg := config.New()
	if cfg == nil {
		mainLogger.Error(ctx, "failed load config")
		return
	}
	//подключамся к постресу
	db, err := postgres.New(cfg.PostgresConfig)
	if err != nil {
		mainLogger.Error(ctx, "failed conn postgres", zap.String("Error:", err.Error()))
		return
	}

	//создаем репозиторий постреса
	rep := repository.NewPostgresRepository(db)
	//подключаемся к Minio
	mo, err := minio.New(ctx, cfg.MinioConfig)
	if err != nil {
		mainLogger.Error(ctx, "failed conn minio client", zap.String("Error:", err.Error()))
		return
	}
	//создаем Minio репозиторий
	mn := repository.NewMinioRepository(mo)
	srv := service.New(rep, mn)
	//создаем грпс сервер
	grpcServer, err := grpc.New(ctx, cfg.GRPCServerPort, srv)
	if err != nil {
		mainLogger.Error(ctx, "failed create new grpc server", zap.String("Error:", err.Error()))
		return
	}
	//запускаем сервер с возможностью плаавной остановки
	graceCh := make(chan os.Signal, 1)
	signal.Notify(graceCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		if err := grpcServer.Start(ctx); err != nil {
			mainLogger.Error(ctx, err.Error())
		}
	}()
	<-graceCh

	grpcServer.Stop(ctx)
}

// migrate -database postgres://root:123@localhost:5432/files?sslmode=disable -path migrations up
// migrate -database postgres://root:123@localhost:5432/files?sslmode=disable -path migrations down
