package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"gitlab.crja72.ru/gospec/go19/messanger/file_service/internal/models"
	"gitlab.crja72.ru/gospec/go19/messanger/file_service/pkg/db/postgres"
	"gitlab.crja72.ru/gospec/go19/messanger/file_service/pkg/logger"
	"gitlab.crja72.ru/gospec/go19/messanger/file_service/pkg/utils"
	"go.uber.org/zap"
)

type PostgresRepository struct {
	db *postgres.DB
}

// Конструктор репозитория постреса
func NewPostgresRepository(db *postgres.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// загружаем данные о файле в базу данных
func (fr *PostgresRepository) Create(ctx context.Context, file *models.File) error {
	err := utils.DoWithTries(func() error {
		_, err := sq.Insert("files").
			Columns("id", "name", "content_type", "public", "sender_id", "recipient_id", "size").
			Values(file.ID, file.Name, file.ContentType, file.Public, file.SenderID, file.RecipientID, file.Size).
			PlaceholderFormat(sq.Dollar).
			RunWith(fr.db.Db).
			Exec()
		if err != nil {
			return fmt.Errorf("%v: %v", ErrFailedInsertNewFileInDb, err)
		}
		return nil
	}, 5, 100*time.Millisecond)
	return err
}

// получаем данные о файле из базы данных
func (fr *PostgresRepository) GetByID(ctx context.Context, file *models.File) error {
	err := utils.DoWithTries(func() error {
		err := sq.Select("name", "content_type", "sender_id", "recipient_id", "size").
			From("files").
			Where(sq.Eq{"id": file.ID}).
			PlaceholderFormat(sq.Dollar).
			RunWith(fr.db.Db).
			QueryRow().
			Scan(&file.Name, &file.ContentType, &file.SenderID, &file.RecipientID, &file.Size)
		if err == sql.ErrNoRows {
			return fmt.Errorf("%v with id=%s", ErrFailedGetFileFromDb, file.ID)
		}
		if err != nil {
			return fmt.Errorf("repository.GetByID: %v", err)
		}
		return nil
	}, 5, 100*time.Millisecond)
	return err
}

// удаляем данные о файле из базы данных
func (fr *PostgresRepository) Delete(ctx context.Context, file *models.File, userid string) (bool, error) {
	err := sq.Select("sender_id", "recipient_id", "public").
		From("files").
		Where(sq.Eq{"id": file.ID}).
		PlaceholderFormat(sq.Dollar).
		RunWith(fr.db.Db).
		QueryRow().
		Scan(&file.SenderID, &file.RecipientID, &file.Public)
	if err == sql.ErrNoRows {
		return false, ErrFailedGetFileFromDb
	}
	if err != nil {
		return false, ErrFailedGetFileFromDb
	}
	if !file.Public {
		if file.SenderID != userid && file.RecipientID != userid {
			logger.New("files").Error(context.Background(), "user have not permisssion to get this file", zap.String("youtId", userid), zap.String("senderId", file.SenderID), zap.String("recipientId", file.RecipientID))
			return false, ErrhaventPermissionToDeleteFile
		}
	}
	err = utils.DoWithTries(func() error {
		res, err := sq.Delete("files").
			Where(sq.Eq{"id": file.ID}).
			PlaceholderFormat(sq.Dollar).
			RunWith(fr.db.Db).
			ExecContext(ctx)
		if err != nil {
			return ErrFailedDeleteFileFromDb
		}
		if a, _ := res.RowsAffected(); a == 0 {
			return ErrFailedDeleteFileFromDb
		}
		return nil
	}, 5, 100*time.Millisecond)
	if err != nil {
		return false, err
	}
	return true, nil
}
