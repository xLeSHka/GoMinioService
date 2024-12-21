package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	"gitlab.crja72.ru/gospec/go19/messanger/file_service/internal/config"
	"gitlab.crja72.ru/gospec/go19/messanger/file_service/internal/repository"
	"gitlab.crja72.ru/gospec/go19/messanger/file_service/internal/service"
	grpc1 "gitlab.crja72.ru/gospec/go19/messanger/file_service/internal/transport/grpc"
	"gitlab.crja72.ru/gospec/go19/messanger/file_service/pkg/db/minio"
	postgres1 "gitlab.crja72.ru/gospec/go19/messanger/file_service/pkg/db/postgres"
	"gitlab.crja72.ru/gospec/go19/messanger/file_service/pkg/logger"
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
	db, err := postgres1.New(cfg.PostgresConfig)
	if err != nil {
		mainLogger.Error(ctx, "failed conn postgres", zap.String("Error:", err.Error()))
		return
	}
	db.Db.DB.Exec(`CREATE TABLE if not EXISTS files
(
    id character varying(64) NOT NULL,
    name character varying(64) NOT NULL,
    content_type character varying(64) NOT NULL,
    public boolean NOT NULL,
    sender_id character varying(64) NOT NULL,
    recipient_id character varying(64),
    size bigint NOT NULL
);`)

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
	grpcServer, err := grpc1.New(ctx, cfg.GRPCServerPort, cfg.CryptoServicePort, srv)
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
