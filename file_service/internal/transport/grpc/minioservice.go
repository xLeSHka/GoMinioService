package grpc

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/bwmarrin/snowflake"
	"github.com/gabriel-vasile/mimetype"
	"gitlab.crja72.ru/gospec/go19/messanger/file_service/internal/models"
	"gitlab.crja72.ru/gospec/go19/messanger/file_service/pkg/api/crypto"
	"gitlab.crja72.ru/gospec/go19/messanger/file_service/pkg/api/file"
	"gitlab.crja72.ru/gospec/go19/messanger/file_service/pkg/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

const (
	snowflakeNode = 1
)

type Service interface {
	UploadFile(ctx context.Context, file *models.File) error
	GetFile(ctx context.Context, file *models.File) error
	DeleteFile(ctx context.Context, file models.File) (bool, error)
}
type FileService struct {
	file.UnimplementedFilesServiceServer
	cryptoClient crypto.CryptoServiceClient
	service      Service
	l            logger.Logger
}

// конструктор сервиса файлов
func NewFileService(srv Service, l logger.Logger, cryptoClient crypto.CryptoServiceClient) *FileService {
	return &FileService{service: srv, l: l, cryptoClient: cryptoClient}
}

// Обрабатываем грпс запрос на загрузку файла
func (s *FileService) UploadFile(ctx context.Context, req *file.UploadFileRequest) (*file.UploadFileResponse, error) {
	var f models.File
	//получаю user id из metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		s.l.Error(context.Background(), "error", zap.Int("code", int(codes.InvalidArgument)),
			zap.String("message", "invalid metadata"))
		return nil, fmt.Errorf("invalid metadata")
	}

	for _, id := range md.Get("X-User-ID") {
		f.SenderID = id
	}
	//проверяю валидность user id
	if f.SenderID == "" {
		s.l.Error(context.Background(), "error", zap.Int("code", int(codes.InvalidArgument)),
			zap.String("message", "invalid sender id"))
		return nil, fmt.Errorf("invalid sender id")
	}

	//генерирую file id
	node, err := snowflake.NewNode(snowflakeNode)
	if err != nil {
		s.l.Error(context.Background(), "error", zap.Int("code", int(codes.Unknown)),
			zap.String("message", err.Error()))
		return nil, err
	}
	f.ID = node.Generate().String()

	f.Name = req.GetName()
	f.Public = req.GetPublic()

	if !f.Public {
		f.RecipientID = req.GetRecipientId()
	}
	if f.Name == "" {
		s.l.Error(context.Background(), "error", zap.Int("code", int(codes.InvalidArgument)),
			zap.String("message", "invalid file name"))
		return nil, fmt.Errorf("invalid file name")
	}
	if f.RecipientID == "" && !f.Public {
		s.l.Error(context.Background(), "error", zap.Int("code", int(codes.InvalidArgument)),
			zap.String("message", "invalid recipient id"))
		return nil, fmt.Errorf("invalid recipient id")
	}

	chunk := req.GetData()

	f.Size += int64(len(chunk))

	if f.Size > 2<<21 {
		s.l.Error(context.Background(), "error", zap.Int("code", int(codes.InvalidArgument)),
			zap.String("message", "file size can not be greater than 4MB"))
		return nil, fmt.Errorf("file size can not be greater than 4MB")
	}
	if f.Size == 0 {
		s.l.Error(context.Background(), "error", zap.Int("code", int(codes.InvalidArgument)),
			zap.String("message", "request have not file data"))
		return nil, fmt.Errorf("request have not file data")
	}
	//считы
	fileType := mimetype.Detect(chunk)
	ext := filepath.Ext(f.Name)
	if ext != fileType.Extension() {

		return nil, fmt.Errorf("file extention does not match the data")
	}
	f.ContentType = fileType.String()

	resp, err := s.cryptoClient.Encrypt(ctx, &crypto.Request{Data: chunk, SecretPhrase: []byte(f.ID)[:8]})

	if err != nil {
		s.l.Error(context.Background(), "error", zap.Int("code", int(codes.Internal)),
			zap.String("message", err.Error()))
		return nil, fmt.Errorf("failed encrypt file data")
	}
	data := resp.GetData()
	s.l.Info(ctx, "data", zap.String("data", string(data)))

	n, err := f.Data.Write(data)
	if err != nil {
		s.l.Error(context.Background(), "error", zap.Int("code", int(codes.Internal)),
			zap.String("message", err.Error()))
		return nil, fmt.Errorf("failed write encrypted file data")
	}
	if n == 0 {
		s.l.Error(context.Background(), "error", zap.Int("code", int(codes.InvalidArgument)),
			zap.Int("n length == 0", n))
		return nil, fmt.Errorf("failed write encrypted file data")
	}
	err = s.service.UploadFile(context.Background(), &f)
	if err != nil {
		s.l.Error(context.Background(), "error", zap.Int("code", int(codes.Unknown)),
			zap.String("message", err.Error()))
		return nil, err
	}

	return &file.UploadFileResponse{Id: f.ID}, nil
}

// Обрабатываем грпс запрос на получение файла
func (s *FileService) GetFile(ctx context.Context, req *file.GetFileRequest) (*file.GetFileResponse, error) {
	var f models.File
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		s.l.Error(context.Background(), "error", zap.Int("code", int(codes.InvalidArgument)),
			zap.String("message", "invalid metadata"))
		return nil, fmt.Errorf("invalid metadata")
	}
	reqId := ""
	for _, id := range md.Get("X-User-ID") {
		reqId = id
	}
	f.ID = req.GetId()
	if f.ID == "" {
		s.l.Error(context.Background(), "error", zap.Int("code", int(codes.InvalidArgument)),
			zap.String("message", "invalid file id"))
		return nil, fmt.Errorf("invalid file id")
	}

	err := s.service.GetFile(context.Background(), &f)
	if err != nil {
		s.l.Error(context.Background(), "error", zap.Int("code", int(codes.Unknown)),
			zap.String("message", err.Error()))
		return nil, err
	}
	if !f.Public {
		if f.SenderID != reqId && f.RecipientID != reqId {
			s.l.Error(context.Background(), "user have not permisssion to get this file", zap.String("youtId", reqId), zap.String("senderId", f.SenderID), zap.String("recipientId", f.RecipientID))
			return nil, fmt.Errorf("iuser have not permission to get this file")
		}
	}
	resp, err := s.cryptoClient.Decrypt(ctx, &crypto.Request{Data: f.Data.Bytes(), SecretPhrase: []byte(f.ID)[:8]})

	if err != nil {
		s.l.Error(context.Background(), "error", zap.Int("code", int(codes.Internal)),
			zap.String("message", err.Error()))
		return nil, err
	}

	return &file.GetFileResponse{Name: f.Name, Data: resp.GetData(), ContentType: f.ContentType, Size: f.Size}, nil
}

// Обрабатываем грпс запрос на удаление файла
func (s *FileService) DeleteFile(ctx context.Context, req *file.DeleteFileRequest) (*file.DeleteFileResponse, error) {
	var f models.File
	f.ID = req.GetId()
	if f.ID == "" {
		s.l.Error(context.Background(), "error", zap.Int("code", int(codes.InvalidArgument)),
			zap.String("message", "invalid file id"))
		return nil, fmt.Errorf("invalid file id")
	}
	ok, err := s.service.DeleteFile(ctx, f)
	if err != nil {
		s.l.Error(context.Background(), "error", zap.Int("code", int(codes.Unknown)),
			zap.String("message", err.Error()))
		return nil, err
	}
	return &file.DeleteFileResponse{Success: ok}, nil
}
