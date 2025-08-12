package minios3_util

import (
	"bytes"
	"context"
	"net/http"
	"net/url"
	"time"

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

func PresignedGetObject(ctx context.Context, bucket, key string, expires time.Duration) (string, error) {
	reqParams := make(url.Values)
	reqParams.Set("response-content-disposition", "attachment")
	presignedURL, err := minioClient.PresignedGetObject(ctx, bucket, key, expires, reqParams)
	if err != nil {
		return "", err
	}
	return presignedURL.String(), nil
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
	putOpts := minio.PutObjectOptions{ContentType: http.DetectContentType(data)}
	_, err = minioClient.PutObject(ctx, bucketKey, fileKey, byteReader, byteReader.Size(), putOpts)
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
	if err != nil {
		return &byteArray, err
	}
	defer reader.Close()

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

// ensurePublicReadPolicy sets a restricted public-read policy allowing access only when
// the HTTP Referer header matches the given domain (e.g., *.port-trak.com).
// Note: Referer can be spoofed or omitted by clients; for stronger control, use a proxy
// or presigned URLs instead.
func ensurePublicReadPolicy(ctx context.Context, bucket string) error {
	domain := "port-trak.com"
	policyJSON := `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Sid": "AllowBucketOpsFromDomain",
				"Effect": "Allow",
				"Principal": "*",
				"Action": ["s3:GetBucketLocation"],
				"Resource": "arn:aws:s3:::` + bucket + `",
				"Condition": {"StringLike": {"aws:Referer": [
					"https://` + domain + `/*",
					"http://` + domain + `/*",
					"https://*.` + domain + `/*",
					"http://*.` + domain + `/*"
				]}}
			},
			{
				"Sid": "AllowGetObjectFromDomain",
				"Effect": "Allow",
				"Principal": "*",
				"Action": ["s3:GetObject"],
				"Resource": "arn:aws:s3:::` + bucket + `/*",
				"Condition": {"StringLike": {"aws:Referer": [
					"https://` + domain + `/*",
					"http://` + domain + `/*",
					"https://*.` + domain + `/*",
					"http://*.` + domain + `/*"
				]}}
			}
		]
	}`

	return minioClient.SetBucketPolicy(ctx, bucket, policyJSON)
}

func Makes3Bucket(namer string) error {
	ctx := context.Background()
	err := minioClient.MakeBucket(ctx, namer, minio.MakeBucketOptions{Region: "us-east-1", ObjectLocking: false})
	if err != nil {
		return err
	}
	// Make the bucket publicly readable so objects are accessible via direct URL
	/*
		if err := ensurePublicReadPolicy(ctx, namer); err != nil {
			return err
		}
	*/
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

func DeleteFileFromS3(bucketKey string, fileKey string) error {
	exists, err := CheckS3BucketExists(bucketKey)
	if err != nil {
		return err
	}

	if !exists {
		return errors.New("bucket does not exist")
	}

	err = minioClient.RemoveObject(context.Background(), bucketKey, fileKey, minio.RemoveObjectOptions{})
	if err != nil {
		return err
	}
	return nil
}
