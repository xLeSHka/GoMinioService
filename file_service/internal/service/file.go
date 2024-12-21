package service

import (
	"context"

	"gitlab.crja72.ru/gospec/go19/messanger/file_service/internal/models"
	"gitlab.crja72.ru/gospec/go19/messanger/file_service/internal/repository"
)

type FileRepo interface {
	Create(ctx context.Context, file *models.File) error
	GetByID(ctx context.Context, file *models.File) error
	Delete(ctx context.Context, file models.File) (bool, error)
}
type Minio interface {
	Create(ctx context.Context, file *models.File) error
	GetByID(ctx context.Context, file *models.File) error
	Delete(ctx context.Context, file models.File) (bool, error)
}
type FileService struct {
	Repo  FileRepo
	Minio Minio
}
//конструктор сервиса
func New(repo FileRepo, minio Minio) *FileService {
	return &FileService{Repo: repo, Minio: minio}
}
//загружаем файл в пострес и минио
func (s *FileService) UploadFile(ctx context.Context, file *models.File) error {

	err := s.Repo.Create(ctx, file)
	if err != nil {
		return err
	}
	err = s.Minio.Create(ctx, file)
	if err != nil {
		return err
	}
	return nil
}
//поулчаем файл из постреса и минио
func (s *FileService) GetFile(ctx context.Context, file *models.File) error {

	err := s.Repo.GetByID(ctx, file)
	if err != nil {
		return err
	}
	err = s.Minio.GetByID(ctx, file)
	if err != nil {
		return err
	}
	return nil
}
//удаляем файл из минио и постреса
func (s *FileService) DeleteFile(ctx context.Context, file models.File) (bool, error) {
	suc, err := s.Repo.Delete(ctx, file)
	if err != repository.ErrFailedDeleteFileFromDb {
		_, err := s.Minio.Delete(ctx, file)
		if err != nil {
			return false, err
		}
	}
	if err != nil {
		return false, err
	}
	suc2, err := s.Minio.Delete(ctx, file)
	if err != nil {
		return false, err
	}
	return suc && suc2, nil
}
