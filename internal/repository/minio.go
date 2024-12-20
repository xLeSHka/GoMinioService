package repository

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/xLeSHka/GoMinioService/internal/models"
	minio1 "github.com/xLeSHka/GoMinioService/pkg/db/minio"
	"github.com/xLeSHka/GoMinioService/pkg/logger"
	"github.com/xLeSHka/GoMinioService/pkg/utils"
	"go.uber.org/zap"
)

type Minio struct {
	mn *minio1.MinioClient
}

// конструктор минио репозитория
func NewMinioRepository(mn *minio1.MinioClient) *Minio {
	return &Minio{mn: mn}
}

// загружаем файл в минио
func (m *Minio) Create(ctx context.Context, file *models.File) error {
	err := utils.DoWithTries(func() error {
		_, err := m.mn.Mc.PutObject(ctx, m.mn.BucketName, file.ID, &file.Data, file.Size, minio.PutObjectOptions{ContentType: file.ContentType})
		if err != nil {
			return fmt.Errorf("%s, name: %s, error:%v", ErrPutMinioObject, file.Name, err)
		}
		return nil
	}, 3, 100*time.Millisecond)

	return err
}

// получаем файл из минио
func (m *Minio) GetByID(ctx context.Context, file *models.File) error {
	err := utils.DoWithTries(func() error {
		obj, err := m.mn.Mc.GetObject(ctx, m.mn.BucketName, file.ID, minio.GetObjectOptions{})
		if err != nil {
			return fmt.Errorf("%s, id: %s, error:%v", ErrGetMinioObject, file.ID, err)
		}
		var fileSize int64 = 0
		buff := make([]byte, 8*1024*256)
		for {
			n, err := obj.Read(buff)
			if err == io.EOF {
				if n > 0 {
					n, _ = file.Data.Write(buff[:n])
					fileSize += int64(n)
				}
				break
			}
			if err != nil {
				return err
			}
			fileSize += int64(n)
			n, _ = file.Data.Write(buff[:n])
		}
		if fileSize != file.Size {
			logger.New("test").Info(ctx, "file", zap.Int64("fileSize", fileSize), zap.Int64("file.Size", file.Size))
			return ErrGetMinioObject
		}
		return nil
	}, 3, 100*time.Millisecond)
	return err
}

// удаляем файл их минио
func (m *Minio) Delete(ctx context.Context, file models.File) (bool, error) {
	err := utils.DoWithTries(func() error {
		err := m.mn.Mc.RemoveObject(ctx, m.mn.BucketName, file.ID, minio.RemoveObjectOptions{})
		if err != nil {
			return fmt.Errorf("%s, id: %s, error:%v", ErrGetMinioObject, file.ID, err)
		}
		return nil

	}, 3, 100*time.Millisecond)
	if err != nil {
		return false, err
	}
	return true, nil
}
