package minios3_util

import (
	"bytes"
	"context"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/pkg/errors"
	"github.com/rogue-syntax/rs-goapiserver/global"
)

var minioClient *minio.Client

func IntiMinioClient() (*minio.Client, error) {
	var err error
	minioClient, err = minio.New(global.EnvVars.MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(global.EnvVars.MinioAccessKey, global.EnvVars.MinioSecretAccessKey, ""),
		Secure: true,
	})
	if err != nil {
		return minioClient, err
	}
	return minioClient, nil
}

func StoreFileToS3(data []byte, bucketKey string, fileKey string) error {

	exists, err := CheckS3BucketExists(bucketKey)
	if err != nil {
		return err
	}

	if !exists {
		err := Makes3Bucket(bucketKey)
		if err != nil {
			return err
		}
	}

	byteReader := bytes.NewReader(data)
	ctx := context.Background()
	_, err = minioClient.PutObject(ctx, bucketKey, fileKey, byteReader, byteReader.Size(), minio.PutObjectOptions{})
	if err != nil {
		return err
	}
	return nil
}

func GetFileFromS3(bucketKey string, fileKey string) (*[]byte, error) {
	var byteArray []byte

	exists, err := CheckS3BucketExists(bucketKey)
	if err != nil {
		return &byteArray, err
	}

	if !exists {
		err := Makes3Bucket(bucketKey)
		if err != nil {
			return &byteArray, err
		}
	}

	reader, err := minioClient.GetObject(context.Background(), bucketKey, fileKey, minio.GetObjectOptions{})
	defer reader.Close()
	if err != nil {
		return &byteArray, err
	}

	stat, _ := reader.Stat()
	byteArray = make([]byte, stat.Size)
	_, err = reader.Read(byteArray)
	if err != nil {
		if err.Error() != "EOF" {
			return &byteArray, err
		}
	}
	return &byteArray, nil

}

func Makes3Bucket(namer string) error {
	err := minioClient.MakeBucket(context.Background(), namer, minio.MakeBucketOptions{Region: "us-east-1", ObjectLocking: false})
	if err != nil {
		return err
	}
	return nil
}

func CheckS3BucketExists(namer string) (bool, error) {
	found, err := minioClient.BucketExists(context.Background(), namer)
	err = errors.Wrap(err, "CheckS3BucketExists")
	if err != nil {
		return found, err
	} else {
		return found, nil
	}
}
