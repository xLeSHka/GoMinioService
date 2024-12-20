package grpc

import (
	"context"
	"io"
	"log"
	"net"
	"os"
	"testing"

	"github.com/xLeSHka/GoMinioService/internal/config"
	"github.com/xLeSHka/GoMinioService/internal/repository"
	"github.com/xLeSHka/GoMinioService/internal/service"
	"github.com/xLeSHka/GoMinioService/pkg/api/file"
	"github.com/xLeSHka/GoMinioService/pkg/db/minio"
	"github.com/xLeSHka/GoMinioService/pkg/db/postgres"
	"github.com/xLeSHka/GoMinioService/pkg/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

// dialer настраивает соединение с grpc серверами
func dialer() func(context.Context, string) (net.Conn, error) {
	listener := bufconn.Listen(1024 * 1024)

	server := grpc.NewServer()

	ctx := context.Background()
	mLogger := logger.New("test")
	ctx = context.WithValue(ctx, logger.LoggerKey, mLogger)
	cfg := config.New()
	if cfg == nil {
		mLogger.Error(context.Background(), "failed load config")
		log.Fatal("failed load config")

	}

	cfg.PostgresConfig.Host = "localhost"
	cfg.MinioConfig.Host = "localhost"

	db, err := postgres.New(cfg.PostgresConfig)
	if err != nil {
		mLogger.Error(ctx, "failed postgres conn", zap.String("Error", err.Error()))
		log.Fatal(err)
	}

	rep := repository.NewPostgresRepository(db)

	minio, err := minio.New(ctx, cfg.MinioConfig)
	if err != nil {
		mLogger.Error(ctx, "failed minio conn", zap.String("Error", err.Error()))
		log.Fatal(err)
	}

	minioRepository := repository.NewMinioRepository(minio)

	srv := service.New(rep, minioRepository)

	file.RegisterFilesServiceServer(server, NewFileService(srv, mLogger))

	go func() {
		if err := server.Serve(listener); err != nil {
			mLogger.Error(ctx, "failed server serve", zap.String("Error", err.Error()))
			log.Fatal(err)
		}
	}()

	return func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}
}
func TestUploadFile(t *testing.T) {
	defer func() {
		if rec := recover(); rec != nil {
			t.Error(rec)
		}
	}()
	type TestUpload struct {
		name        string
		fileName    string
		filePublic  bool
		recipientId string
		err         string
	}
	Tests := []TestUpload{
		{
			name:        "upload .txt file",
			fileName:    "test.txt",
			filePublic:  false,
			recipientId: "upload_test",
		},
		{
			name:       "upload .webm file",
			fileName:   "golang-golang-halloween.webm",
			filePublic: true,
		},
		{
			name:        "bad file name",
			fileName:    "",
			filePublic:  false,
			recipientId: "upload_test",
			err:         "invalid file name",
		},
		{
			name:        "bad recipient id",
			fileName:    "failRecipient.txt",
			filePublic:  false,
			recipientId: "",
			err:         "invalid recipient id",
		},
		{
			name:       "bad file size",
			fileName:   "golang-golang-halloween.gif",
			filePublic: true,
			err:        "file size can not be greater than 4MB",
		},
		{
			name:        "no data",
			fileName:    "failNoData.txt",
			filePublic:  false,
			recipientId: "upload_test",
			err:         "request have not file data",
		},
	}

	ctx := context.Background()
	mLogger := logger.New("test")
	ctx = context.WithValue(ctx, logger.LoggerKey, mLogger)

	conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer()))
	if err != nil {
		t.Error(err)
	}
	defer conn.Close()

	client := file.NewFilesServiceClient(conn)

	for _, test := range Tests {
		t.Run(test.name, func(t *testing.T) {

			if test.fileName == "" {
				_, err := client.UploadFile(ctx, &file.UploadFileRequest{Name: test.fileName, Public: test.filePublic, RecipientId: test.recipientId})
				if err != nil {
					if status.Convert(err).Message() != test.err {
						t.Errorf("error message: expected %v, recieved %v", test.err, status.Convert(err).Message())
					}
				}
				return
			}

			f, err := os.OpenFile("./testdata/"+test.fileName, os.O_RDONLY, 0600)
			if err != nil {
				t.Error(err)
				return
			}

			data, err := io.ReadAll(f)

			if err != nil {
				t.Error(err)
				return
			}
			resp, err := client.UploadFile(ctx, &file.UploadFileRequest{
				Name:        test.fileName,
				Public:      test.filePublic,
				RecipientId: test.recipientId,
				Data:        data,
			})

			if err != nil {
				if status.Convert(err).Message() != test.err {
					t.Errorf("error message: expected %v, recieved %v", test.err, status.Convert(err).Message())
				}
				return
			}

			if resp != nil {
				if resp.Id == "" {
					t.Error("failed return file id")
				}
				return
			}
			return
		})
	}

}

func TestGetFile(t *testing.T) {
	defer func() {
		if rec := recover(); rec != nil {
			t.Error(rec)
		}
	}()
	type TestGet struct {
		name        string
		fileName    string
		filePublic  bool
		recipientId string
		id          string
		err         string
	}
	Tests := []TestGet{
		{
			name:        "get .txt file",
			fileName:    "test.txt",
			filePublic:  false,
			recipientId: "recipitne",
			id:          "1",
		},
		{
			name:       "get .webm file",
			filePublic: true,
			fileName:   "golang-golang-halloween.webm",
			id:         "2",
		},

		{
			name: "get file with bad  id",
			id:   "",
			err:  "invalid file id",
		},
	}

	ctx := context.Background()
	mLogger := logger.New("test")
	ctx = context.WithValue(ctx, logger.LoggerKey, mLogger)

	conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer()))
	if err != nil {
		t.Error(err)
	}
	defer conn.Close()

	client := file.NewFilesServiceClient(conn)

	for _, test := range Tests {
		t.Run(test.name, func(t *testing.T) {
			// test with empty id
			if test.id == "" {
				_, err := client.GetFile(ctx, &file.GetFileRequest{Id: test.id})

				if err != nil {
					if status.Convert(err).Message() != test.err {
						t.Errorf("error message: expected %v, recieved %v", test.err, status.Convert(err).Message())
					}
				}
				return
			}
			//upload file for get file id
			f, err := os.OpenFile("./testdata/"+test.fileName, os.O_RDONLY, 060)
			if err != nil {
				t.Error(err)
				return
			}

			data, err := io.ReadAll(f)
			if err != nil {
				t.Error(err)
				return
			}

			resp, err := client.UploadFile(ctx, &file.UploadFileRequest{
				Name:        test.fileName,
				Public:      test.filePublic,
				RecipientId: test.recipientId,
				Data:        data,
			})
			if err != nil {
				t.Error(err)
				return
			}

			getResp, err := client.GetFile(ctx, &file.GetFileRequest{
				Id: resp.Id,
			})
			if err != nil {
				if status.Convert(err).Message() != test.err {
					t.Errorf("error message: expected %v, recieved %v", test.err, status.Convert(err).Message())
					return
				}
			}
			if getResp != nil {
				if getResp.Size != int64(len(data)) {
					t.Errorf("recieved file size not match expected file size")
					return
				}
			}
		})
	}
}

func TestDeleteFile(t *testing.T) {
	type TestDelete struct {
		name        string
		fileName    string
		filePublic  bool
		recipientId string
		id          string
		err         string
	}
	Tests := []TestDelete{
		{
			name:        "delete .txt file",
			fileName:    "test.txt",
			filePublic:  false,
			recipientId: "recipitne",
			id:          "1",
		},

		{
			name:       "delete .webm file",
			filePublic: true,
			fileName:   "golang-golang-halloween.webm",
			id:         "2",
		},
		{
			name: "delete file with bad id",
			id:   "",
			err:  "invalid file id",
		},
	}
	ctx := context.Background()
	mlogger := logger.New("test")
	ctx = context.WithValue(ctx, logger.LoggerKey, mlogger)

	conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer()))
	if err != nil {
		t.Error(err)
		return
	}
	defer conn.Close()

	client := file.NewFilesServiceClient(conn)
	for _, test := range Tests {
		t.Run(test.name, func(t *testing.T) {
			//delete file with empty id
			if test.id == "" {
				_, err := client.DeleteFile(ctx, &file.DeleteFileRequest{Id: test.id})
				if err != nil {
					if status.Convert(err).Message() != test.err {
						t.Errorf("error message: expected %v, recieved %v", status.Code(err), test.err)
					}

					return
				}
			}

			f, err := os.OpenFile("./testdata/"+test.fileName, os.O_RDONLY, 0600)
			if err != nil {
				t.Error(err)
				return
			}
			data, err := io.ReadAll(f)
			if err != nil {
				t.Error(err)
				return
			}
			resp, err := client.UploadFile(ctx, &file.UploadFileRequest{
				Name:        test.fileName,
				Public:      test.filePublic,
				RecipientId: test.recipientId,
				Data:        data,
			})
			if err != nil {
				t.Error(err)
				return
			}

			//delete file

			delResp, err := client.DeleteFile(ctx, &file.DeleteFileRequest{Id: resp.Id})
			if err != nil {
				t.Error(err)
				return
			}
			if delResp.Success != true {
				t.Errorf("resp status: expected %v, recieved %v", true, delResp.Success)
				return
			}

		})
	}
}
