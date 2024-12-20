package minio

import (
	"context"
	"fmt"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/xLeSHka/GoMinioService/pkg/logger"
)

type MinioConfig struct {
	Host       string `env:"MINIO_HOST" env-default:"minio"`
	Port       int    `env:"MINIO_PORT" env-default:"9000"`
	BucketName string `env:"MINIO_BUCKET_NAME" env-default:"bucket"`
	User       string `env:"MINIO_ROOT_USER" env-default:"fileservice"`
	Password   string `env:"MINIO_ROOT_PASSWORD" env-default:"minio123"`
}
type MinioClient struct {
	Mc         *minio.Client
	BucketName string
}

// Подключение к минио
func New(ctx context.Context, config MinioConfig) (*MinioClient, error) {
	endpoint := fmt.Sprintf("%s:%d", config.Host, config.Port)
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.User, config.Password, ""),
		Secure: false,
	})
	if err != nil {
		return nil, err
	}
	mc := &MinioClient{
		Mc:         client,
		BucketName: config.BucketName,
	}
	exist, _ := mc.Mc.BucketExists(ctx, config.BucketName)
	logger.GetLoggerFromCtx(ctx).Info(ctx, config.BucketName)
	if !exist {
		err := mc.Mc.MakeBucket(ctx, config.BucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, err
		}
	}
	return mc, nil
}
