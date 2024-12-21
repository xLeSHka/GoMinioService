package repository

import "errors"

var (
	ErrFailedInsertNewFileInDb = errors.New("failed insert new file in db")
	ErrFailedGetFileFromDb     = errors.New("failed select file")
	ErrFailedDeleteFileFromDb  = errors.New("failed delete file")
	ErrPutMinioObject          = errors.New("failed put object to minio")
	ErrGetMinioObject          = errors.New("failed get object from minio")
	ErrhaventPermissionToDeleteFile = errors.New("you havent permission to delete this file")
)
